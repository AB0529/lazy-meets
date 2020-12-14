import calendar
import time
from datetime import datetime, timedelta
from threading import Thread

from pynput.keyboard import Controller
from selenium import webdriver
from selenium.webdriver.firefox.options import Options
from termcolor import colored

from settings import config


# Custom error function
def error(msg):
    print(f"[{ colored('ERROR', 'red') }] - {msg}")


# Join class function
def join_class(name, url):
    print(f"[{colored('INFO', 'blue')}] - joining class {colored(name, 'yellow')}")

    # Browser options
    options = Options()
    if config.HEADLESS:
        options.add_argument("--headless")
    options.add_argument("--disable-infobars")
    options.set_preference("permissions.default.microphone", 2)
    options.set_preference("permissions.default.camera", 2)

    # Browser
    browser = webdriver.Firefox(options=options)
    browser.set_window_size(512, 512)
    browser.get(
        "https://accounts.google.com/ServiceLogin?service=mail&continue=https://mail.google.com/mail/#identifier")

    # Fill in email
    email = browser.find_element_by_id("identifierId")
    email.send_keys(config.EMAIL)
    nextButton = browser.find_element_by_id("identifierNext")
    nextButton.click()
    time.sleep(3)
    # Fill in password
    Controller()
    password = browser.find_element_by_xpath("//input[@class='whsOnd zHQkBf']")
    password.send_keys(config.PASSWORD)
    signInButton = browser.find_element_by_id("passwordNext")
    signInButton.click()
    time.sleep(3)

    # Go to meet
    browser.get(url)
    # Wait for it to load
    time.sleep(10)
    # Join button
    browser.find_element_by_xpath(
        "//span[contains(text(), 'Join')]"
    ).click()
    time.sleep(10)

    # Number of people in call
    # TODO: This is ugly make it better
    num = int(
        browser.find_element_by_xpath(
            "//span[@class='NPEfkd RveJvd snByac']"
        ).text
    )
    print(
        f"[{colored('INFO', 'blue')}] - Found {colored('yellow', num)} people in the class"
    )

    while True:
        num = int(
            browser.find_element_by_xpath(
                "//span[@class='NPEfkd RveJvd snByac']"
            ).text
        )
        if num < config.LEAVE_THRESHOLD:
            break

        # Check every 10 seconds
        time.sleep(10)

    browser.quit()


# Checks if time is in a range
def time_in_range(start, end, now):
    if start < end:
        return now >= start and now <= end
    else:
        return now >= start or now <= end


# Main function
def main(now):
    # Gets the current class based on the weekday
    weekday = calendar.day_name[now.weekday()].lower()

    # Make sure there's a class for the day
    if not any(weekday in wd.get("weekday") for wd in config.SCHEDULE):
        error(f"Schedule has no day {colored(weekday, 'yellow')}")
        return False

    # Get the classes for the day
    classes = [
        c.get("classes") for c in config.SCHEDULE if weekday in c.get("weekday")
    ][0]

    # Check the time for each class with current time
    for c in classes:
        # Time to join class
        jt = datetime.strptime(
            f'{now.year}-{now.month}-{now.day} {c.get("join_time")}',
            "%Y-%m-%d %H:%M").replace(second=0, microsecond=0)

        # If matches with current time, join class
        if time_in_range(jt, jt + timedelta(minutes=config.SKIP_THRESHOLD), now):
            class_thread = Thread(target=join_class(
                c.get("name"), c.get("url")))
            # Don't join a class if already in a class
            if class_thread.is_alive():
                return


# Main loop
while True:
    now = datetime.now().replace(second=0, microsecond=0)
    main(now)

    # Wait 1 min before checking again
    time.sleep(60)

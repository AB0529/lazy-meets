# Lazy-Meets

Are you so lazy you can't even go to online classes? Well fret no more!

This uses Selenium with Firefox to automate loggin in and going to your online classes on a fixed schedule.

You can leave this running in the background forever and it will automaticlly start a Firefox instance and join a class when the time is right.

For leaving a class, it will check for a threshold of users are left, once the amount of users are below this threshold, it will leave the meet by quitting the browser.

To change this threshold and edit other details, look in `settings/template_config.py`. Make sure to `rename the file to config.py`!

# Requirements

* Python 3 or higher
* Selenium with Gecko Driver ([Install Guide](https://selenium-python.readthedocs.io/installation.html))

# Installation

## Install dependencies via pip
```sh
pip3 install termcolor selenium
```
## Setup config
* Rename `template_config.py` in the `settings` folder to `config.py`
* Open the file and edit it to your schedule and account detail needs
# Usage

```sh
python3 index.py
```
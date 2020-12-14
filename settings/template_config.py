# Will leave the call when below this number of people are left in the call
LEAVE_THRESHOLD = 10
# Change to true to no longer see a browser window open
HEADLESS = False
# The window of time to join a class even if the join time has passed
# For example, if it was 5, it will join a class if the current time is 7:28 and the join time is 7:23
SKIP_THRESHOLD = 5  # Minutes

# Too lazy to add, will do it later, ignore this
# ----------------------------------------------------------
# Minutes to wait if failed to join class
# RETRY_INTERVAL = 1
# How many times to retry joining until giving up
# RETRY_THRESHOLD = 3
# ----------------------------------------------------------

# The shedule for the week
# I recommend you join the class 3 minutes late
# So you avoid awkward moments when it's only you
# and the teacher in the call :)
SCHEDULE = [
    # Monday and wednesday classes
    {
        "weekday": ["monday", "wednesday"],
        "classes": [
            {
                "name": "Period 1",
                "url": "",
                "join_time": "7:23"  # !!! THIS IS IN 24-hour TIME FORMAT !!!
            },
            # {
            #     "name": "Period 2",
            #     "url": "",
            #     "join_time": "8:43"
            # },
        ]
    },
    # Tuesday and thursday classes
    # {
    #     "weekday": ["tuesday", "thursday"],
    #     "classes": [
    #         {
    #             "name": "Period 4",
    #             "url": "URL HERE",
    #             "join_time": "8:43"
    #         },
    #     ]
    # },
]

# Authentication
EMAIL = "EMAIL"
PASSWORD = "PASSWORD"

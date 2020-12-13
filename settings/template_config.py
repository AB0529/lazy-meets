# Will leave the call when this number or below of people are left in the call
LEAVE_THRESHOLD = 10

# Too lazy to add, will do it later, ignore this
# ----------------------------------------------------------
# Minutes to wait if failed to join class
# RETRY_INTERVAL = 1
# How many times to retry joining until quitting
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
                "join_time": "7:23"
                },
                # {
                #     "name": "Period 2",
                #     "url": "",
                #     "join_time": "8:43"
                # },
                # {
                #     "name": "Period 3",
                #     "url": "",
                #     "join_time": "11:16"
                # },
                # {
                #     "name": "Period 6",
                #     "url": "",
                #     "join_time": "12:36"
                # }
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
    #         {
    #             "name": "Period 5",
    #             "url": "URL",
    #             "join_time": "11:16"
    #         }
    #     ]
    # },
]

# Authentication
EMAIL = "EMAIL"
PASSWORD = "PASSWORD"
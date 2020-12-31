# Reminders Application

This application is written in golang following a tutorial by `@steevehook (Steve Hook)` and his git repo reference, it's a cli based application. The sole purpose of this repo and code is to learn golang to build
applications

The following are the commands to run and use the application

## Components 🧩
- CLI Client
- HTTP client for communicating with the Backend API
- Backend API
- HTTP client for communicating with the Notifier service
- Notifier service
- Background Saver worker
- Background Notifier worker
- JSON file Database (db.json)
- Database config file (.db.config.json)

# Functionality 
* create a reminder
* edit a reminder
* fetch a list of reminders
* delete a list of reminders

### make commands
1) `make` - builds the client and the server
2) `make client` - builds the client binary
3) `make server` - builds the server binary

# server flags
```
# display a helpful message of all available flags for the server binary
./bin/server --help

# runs the backend http server on the specified address
# --backend flag needs to be provided to ./bin/client if address != :8080
./bin/server --addr=":9090"

# runs the http backend server with a different path to the database
./bin/server --db="/path/to/db.json"

# runs the http backend server with a different path to the database config
./bin/server --db-cfg="/path/to/.db.config.json"

# runs the http backend server with a different notifier service url
./bin/server --notifier="http://localhost:8989"
```
# Client flags

```


# displays a helpful message about all the commands and flags available
./bin/client --help

# runs CLI client with a different backend api url
./bin/client --backend="http://localhost:7777"

# creates a new reminder which will be notified after 3 minutes
./bin/client create --title="Some title" --message="Some msg!" --duration=3m

# edits the reminder with id: 13
# note: if the duration is edited, the reminder gets notified again
./bin/client edit --id=13 --title="Another title" --message="Another msg!"

# fetches a list of reminders with the following ids
./bin/client fetch --id=1 --id=3 --id=6

# deleted the reminders with the following ids
./bin/client delete --id=2 --id=4

```
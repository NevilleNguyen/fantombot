# Fantom bot

Used to track events on Fantom chain.

# How it works
You have to update these following configures in `env/mainnet.json` config file:
- `min_staking_amount`
- `min_claim_amount`
- `min_transfer_amount`
- `telegram` fields `tokens` and `chat_id`

# Builds
Run command:
```go build -o ./build/fantombot```

# Run 
Run command:
```./build/fantombot start```

# Run as Ubuntu service
Firstly take a look at the file `fantombot.service`, you have to update the following fields:
- `ConditionPathExists`
- `User`
- `Group`
- `WorkingDirectory`
- `ExecStart`

Then copy the file `fantombot.service` to the path `/etc/systemd/system/`. Reload the system daemon by running 
`sudo systemctl daemon-reload`.

 After that you can start the service and check the status by using:
```
systemctl fantom.service start
systemctl fantom.service status
```

Or watching logs
```
journalctl -u fantombot -f
```


## Scout - IP Discord Bot

This is a Discord bot that allows users to retrieve the server's public IP address using the `/get-ip` command.

### Prerequisites

* A Discord bot token
* Go installed on your system ([https://go.dev/doc/install](https://go.dev/doc/install))

### Running the bot

1. Build the bot executable using:

```
go build
```

2. Run the bot executable with the following flag:

```
./scout --token=<YOUR_BOT_TOKEN>
```

**Replace `<YOUR_BOT_TOKEN>` with your actual Discord bot token.**

**Example:**

```
./scout --token=your_bot_token_here
```

### Usage

In your Discord server, type the following command to get the server's public IP address:

```
/get-ip
```

The bot will respond with the retrieved IP address in a message.

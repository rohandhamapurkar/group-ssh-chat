# Group SSH Chat

A lightweight and secure group chat system that operates over SSH connections, allowing users to communicate through terminal sessions.

## Overview

Group SSH Chat creates a simple chat server that users can connect to using standard SSH clients. It leverages SSH's built-in public key authentication for secure access control. When a user connects, they can send messages that are broadcast to all other connected users in real-time.

## Features

- **SSH-based Authentication**: Securely authenticate users using SSH public key authentication
- **Real-time Group Chat**: All connected users can communicate with each other
- **Enhanced Terminal UI**: Rich terminal interface featuring:
  - Color-coded usernames and timestamps
  - ANSI color support for better visual distinction
  - Real-time user list updates
  - Clear visual separators between different sections
- **Interactive Commands**: Support for slash commands:
  - `/help` - Shows a list of available commands
  - `/users` - Shows a list of online users
  - `/whisper <username> <message>` - Sends a private message to another user
  - `/clear` - Clears the chat screen
- **Private Messaging**: Send private messages to specific users
- **Multiple Concurrent Sessions**: Support for multiple users connecting simultaneously
- **Session Management**: Tracks active client sessions with unique IDs

## Requirements

- Go 1.20 or higher
- SSH keys for server and clients
- Linux/macOS/Windows with SSH client support
- Terminal with Unicode and color support

## Installation

1. Clone the repository:
   ```bash
   git clone https://github.com/rohandhamapurkar/group-ssh-chat.git
   cd group-ssh-chat
   ```

2. Install dependencies:
   ```bash
   go mod download
   ```

3. Create a `.env` file with the following settings:
   ```
   SSH_SERVER_HOST=0.0.0.0
   SSH_SERVER_PORT=2222
   HOST_SSH_PRIVATE_KEY_PATH=/path/to/your/server_private_key
   AUTHORIZED_KEYS_PATH=/path/to/your/authorized_keys
   ```

4. Build the application:
   ```bash
   # On Linux/macOS
   go build -o group-ssh-chat cmd/main.go

   # On Windows
   go build -o group-ssh-chat.exe cmd/main.go
   ```

## Setup

### Server SSH Keys

1. Generate an SSH key pair for the server if you don't have one:
   ```bash
   ssh-keygen -t rsa -b 4096 -f server_key
   ```

2. Update the `.env` file with the path to your private key:
   ```
   HOST_SSH_PRIVATE_KEY_PATH=./server_key
   ```

### Authorized Keys

1. Create an authorized_keys file with public keys for all users:
   ```bash
   touch authorized_keys
   ```

2. Add user public keys to this file in the standard SSH format, but with usernames as comments:
   ```
   ssh-rsa AAAAB3NzaC1yc2E... username1
   ssh-rsa AAAAB3NzaC1yc2E... username2
   ```

3. Update the `.env` file with the path to your authorized_keys file:
   ```
   AUTHORIZED_KEYS_PATH=./authorized_keys
   ```

## Usage

### Starting the Server

Run the server:
```bash
go run cmd/main.go
```

### Connecting as a Client

Users can connect using any standard SSH client:
```bash
ssh -p 2222 username@server_address -i /path/to/user_private_key
```

Once connected, you'll see a terminal interface with chat messages and a user list. Type your message and press Enter to send it to all connected users. You can also use slash commands like `/help`, `/users`, `/whisper`, and `/clear` for additional functionality.

## Architecture

The project consists of five main components:

1. **auth**: Handles SSH authentication mechanisms, including public key verification
2. **sshserver**: Manages the SSH server, including client connections and message broadcasting
3. **ui**: Provides the terminal user interface with ANSI color support
4. **commands**: Implements the slash command system for interactive features
5. **main**: Initializes the application and starts the server

## Security Considerations

- Password authentication is disabled by default (commented out in the code)
- All communication is encrypted via SSH
- Users are authenticated using public key authentication
- The server requires proper setup of SSH keys and authorized_keys file

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

```
MIT License

Copyright (c) 2025 Rohan Dhamapurkar

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
```

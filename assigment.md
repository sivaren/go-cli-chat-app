**Chat Application**: Create a simple chat application using
WebSockets, allowing users to send and receive messages
in **real-time**

- Create CLI (command-line interface) for chat application
- Implement WebSocket connections for a real-time communication between
    client and server
- Function to send messages to all users in a chat room
- DM (Direct message)
- User authentication
- Allow users to join and leave chat rooms
- Allow users to create chat rooms
- Store messages (Database, File, etc.)
- Implement error handling for scenarios such as connection error
- **Docker/Docker-compose** for deployment

Note:
- Go
    - You can use the Gorilla WebSocket library to implement WebSockets.
    - Focus on writing clean and readable code, idiomatic Go code with proper error handling, concurrency, and documentation.
- Rust
    - You can use Axum/tokio-tungstenite/Rocket to implement WebSockets.
    - You can use thiserror/anyhow library to implement error handling in your chat application
    - Focus on writing clean and readable code, idiomatic Rust code with proper error handling, concurrency, and documentation.
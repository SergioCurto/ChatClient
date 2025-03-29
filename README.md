# ChatClient

ChatClient is a Go application that connects to multiple chat platforms (currently Twitch and YouTube) and aggregates the messages into a single, unified view. This allows you to monitor multiple chat sources simultaneously without having to switch between different applications or browser tabs.

## Code structure

```
github.com/SergioCurto/ChatClient/
├── cmd/
│   └── chat_client/              # Executable application
│       └── main.go               
├── internal/                     # Internal application code (not meant to be imported by external projects)
│   ├── aggregator/               # Core logic for aggregating chat messages
│   │   ├── aggregator.go         
│   │   └── aggregator_test.go    
│   ├── chatconsumers/            
│   │   ├── console/              # Console chat consumer
│   │   │   └── console.go        
│   │   ├── simplepage/           # Simple page chat consumer
│   │   │   └── simplepage.go     
│   │   ├── chatconsumer.go       # Interface for chat consumers
│   │   └── chatconsumer_test.go  
│   ├── chatproviders/            
│   │   ├── twitch/               # Twitch chat provider
│   │   │   └── twitch.go         
│   │   ├── youtube/              # Youtube chat provider
│   │   │   └── youtube.go        
│   │   ├── chatprovider.go       # Interface for chat providers
│   │   └── chatprovider_test.go  
│   ├── chatmodels/               
│   │   └── chatmessage.go        # Structure for chat message
│   └── config/                   
│       └── config.go             # Configuration management and environment file loading
├── .env                          # Environment variables (do not commit this file to version control)
├── .env.example                  # Example of environment variables (use this to create your .env file)
├── go.mod                        # Go module definition
├── go.sum                        # Go module checksums
└── README.md                     # Project documentation
```

Chat providers implement the `ChatProvider` interface and chat consumers implement the `ChatConsumer` interface, these are then used by the `Aggregator`. 

Messages are published by the providers and forwarded to the consumers by `Aggregator` using Go channels, with a simplified publish-subscribe pattern.

ChatProviders and ChatConsumers are created using a factory pattern, allowing for easy extension with new providers and consumers. The factory pattern also allows for the creation of multiple instances of the same provider or consumer with different configurations if needed.

Go routines are used to concurrently collect messages from different chat providers and to process messages by the consumers. Wait groups are used to manage the lifecycle of the providers and consumers, ensuring that all active providers are gracefully disconnected and all consumers have finished processing before the application terminates.


## Configuration

The application is configured using environment variables. Create a `.env` file in the project root directory (based on `.env.example`). At least one chat provider and one chat consumer needs to be enabled for the application to work.

Chat consumers:
- Console output: `OUTPUT_CHAT=true`
- Simple page output: `OUTPUT_SIMPLEPAGE=true` (default page http://localhost:8080)

Chat providers:
- Twitch: `CONNECT_TWITCH=true`
- Youtube: `CONNECT_YOUTUBE=true`

**Required if `CONNECT_TWITCH=true`:**

*   `TWITCH_CHANNEL`: Twitch channel to connect to (e.g., `your_twitch_channel`).

**Required if `CONNECT_YOUTUBE=true`:**

*   `YOUTUBE_CHANNEL_ID`: YouTube channel id to connect to (e.g., `your_youtube_channel`, do not confuse with a channel handle `@channel_handle`).
*   `YOUTUBE_API_KEY`: Api key used to connect to the Youtube API (create it on https://console.cloud.google.com/apis/api/youtube.googleapis.com/credentials)
*   `YOUTUBE_QUERIES_PER_DAY`: Number of queries per day allowed to the Youtube API (default: `10000`)

## Executing

## Running the Application

1.  **Configuration:** Ensure you have created a `.env` file and configured the necessary environment variables (see the "Configuration" section)
2.  **Build (Optional):** To build a standalone executable, run:
    ```bash
    go build ./cmd/chat_client/main.go
    ```
    This will create an executable file named `main` (or `main.exe` on Windows) in the project root.
3.  **Run:**
    *   **From source:** To run directly from the source code, use:
        ```bash
        go run ./cmd/chat_client/main.go
        ```
    *   **From executable:** To run the built executable, use:
        ```bash
        ./main
        ```
        (or `.\main.exe` on Windows)


## Running the tests

To run the tests, navigate to the root directory of the project and execute the following command:

```bash
go test ./...
```

import React, { useEffect, useState } from "react";

export default function LiveLogs() {
  const [logs, setLogs] = useState([]);

  useEffect(() => {
    // Create the WebSocket connection
    const ws = new WebSocket("ws://localhost:8080/ws");

    // Called when the WebSocket connection is established
    ws.onopen = () => {
      console.log("WebSocket connection established to /ws");
    };

    // Called whenever a message is received from the server
    ws.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data);
        // Prepend the new log to the list, keeping only the last 50
        setLogs((prevLogs) => [data, ...prevLogs].slice(0, 50));
      } catch (error) {
        console.error("Error parsing WebSocket message:", error);
      }
    };

    // Called when an error occurs
    ws.onerror = (error) => {
      console.error("WebSocket error:", error);
    };

    // Clean up the connection when the component unmounts
    return () => {
      ws.close();
    };
  }, []);

  return (
    <div className="card mt-4">
      <h2>Live Logs</h2>
      <div className="scrollable-list">
        <ul>
          {logs.map((log, index) => (
            <li key={index}>
              <strong>[{log.log_level || "N/A"}] {log.source || "Unknown"}:</strong>{" "}
              {log.message || "No message"}
            </li>
          ))}
        </ul>
      </div>
    </div>
  );
}

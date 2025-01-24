import React from "react";
import LogQuery from "./components/LogQuery";
import Metrics from "./components/Metrics";
import LiveLogs from "./components/LiveLogs";

export default function App() {
  return (
    <div className="dashboard-container">
      <header>
        <h1>Distributed Logging Dashboard</h1>
        <p>View and analyze logs in real-time</p>
      </header>
      <main>
        <div className="grid grid-2-cols">
          <LogQuery />
          <Metrics />
        </div>
        <LiveLogs />
      </main>
    </div>
  );
}

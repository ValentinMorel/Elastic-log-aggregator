import React, { useState, useEffect } from "react";
import axios from "axios";

export default function Metrics() {
  const [metrics, setMetrics] = useState({
    activeSources: 0,
    alertsTriggered: 0,
  });

  useEffect(() => {
    let intervalId;
  
    const fetchMetrics = async () => {
      try {
        const response = await axios.get("http://localhost:8080/metrics");
        setMetrics(response.data);
      } catch (error) {
        console.error("Error fetching metrics:", error);
      }
    };
  
    // Fetch once immediately
    fetchMetrics();
  
    // Then fetch periodically (e.g., every 30 seconds)
    intervalId = setInterval(fetchMetrics, 3000);
  
    // Cleanup on unmount
    return () => clearInterval(intervalId);
  }, []);
  

  return (
    <div className="card">
      <h2>System Metrics</h2>
      <p>Active Sources: {metrics.activeSources}</p>
      <p>Alerts Triggered: {metrics.alertsTriggered}</p>
    </div>
  );
}

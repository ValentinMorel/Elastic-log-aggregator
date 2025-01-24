import React, { useState } from "react";
import axios from "axios";

export default function LogQuery() {
  const [filters, setFilters] = useState({
    source: "",
    logLevel: "",
    startTime: "",
    endTime: "",
  });
  const [logs, setLogs] = useState([]);

  const handleChange = (e) => {
    setFilters({ ...filters, [e.target.name]: e.target.value });
  };

  // Build an Elasticsearch query object from user filters
  const buildEsQuery = () => {
    const must = [];

    // Match on exact source (if provided)
    if (filters.source) {
      must.push({ term: { source: filters.source } });
    }

    // Match on log level (if provided)
    if (filters.logLevel) {
      must.push({ term: { log_level: filters.logLevel } });
    }

    // Filter by time range (if provided)
    // Converts yyyy-MM-ddTHH:mm -> ISO8601 for Elasticsearch
    if (filters.startTime && filters.endTime) {
      must.push({
        range: {
          timestamp: {
            gte: new Date(filters.startTime).toISOString(),
            lte: new Date(filters.endTime).toISOString(),
          },
        },
      });
    }

    // If no filters, do match_all
    if (must.length === 0) {
      return { match_all: {} };
    }
    return { bool: { must } };
  };

  const handleSearch = async (e) => {
    e.preventDefault();
    try {
      // Build the request body
      const esQuery = {
        query: buildEsQuery(),
        sort: [{ timestamp: { order: "desc" } }],  // Sort descending by timestamp
        size: 50,                                  // Limit to 50 results
      };

      // POST to Elasticsearch with JSON
      const response = await axios.post(
        "http://localhost:9200/logs/_search?pretty",
        esQuery,
        {
          headers: { "Content-Type": "application/json" },
        }
      );

      // Elasticsearch returns documents in response.data.hits.hits
      const hits = response.data?.hits?.hits || [];
      // Each hit has an _source object with the actual log fields
      const formattedLogs = hits.map((hit) => hit._source);

      setLogs(formattedLogs);
    } catch (error) {
      console.error("Error fetching logs:", error);
    }
  };

  return (
    <div className="card">
      <h2>Query Logs</h2>
      <form onSubmit={handleSearch}>
        <label htmlFor="source">Source</label>
        <input
          type="text"
          id="source"
          name="source"
          value={filters.source}
          onChange={handleChange}
        />

        <label htmlFor="logLevel">Log Level</label>
        <select
          id="logLevel"
          name="logLevel"
          value={filters.logLevel}
          onChange={handleChange}
        >
          <option value="">Any</option>
          <option value="INFO">INFO</option>
          <option value="WARN">WARN</option>
          <option value="ERROR">ERROR</option>
        </select>

        <label htmlFor="startTime">Start Time</label>
        <input
          type="datetime-local"
          id="startTime"
          name="startTime"
          value={filters.startTime}
          onChange={handleChange}
        />

        <label htmlFor="endTime">End Time</label>
        <input
          type="datetime-local"
          id="endTime"
          name="endTime"
          value={filters.endTime}
          onChange={handleChange}
        />

        <button type="submit" className="button">
          Search
        </button>
      </form>

      <div className="mt-4">
        <h3 className="mb-2">Results:</h3>
        <div className="scrollable-list">
          <ul>
            {logs.map((log, index) => (
              <li key={index}>
                <strong>
                  [{log.log_level || "N/A"}] {log.source || "Unknown"}:
                </strong>{" "}
                {log.message || "No message"}
              </li>
            ))}
          </ul>
        </div>
      </div>
    </div>
  );
}

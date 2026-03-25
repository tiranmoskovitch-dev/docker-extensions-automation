import React, { useEffect, useState } from "react";
import { createDockerDesktopClient } from "@docker/extension-api-client";

const ddClient = createDockerDesktopClient();

interface PortMapping {
  container: string;
  service: string;
  port: number;
  protocol: string;
  status: "open" | "closed" | "conflict";
  stack: string;
}

export function App() {
  const [ports, setPorts] = useState<PortMapping[]>([]);
  const [loading, setLoading] = useState(true);

  const fetchPorts = async () => {
    try {
      const containers = await ddClient.docker.listContainers() as any[];
      const portMap: PortMapping[] = [];
      const seen = new Map<number, string>();

      for (const c of containers) {
        const stack = c.Labels?.["com.docker.compose.project"] || "--";
        const service = c.Labels?.["com.docker.compose.service"] || c.Names?.[0]?.replace("/", "") || c.Id?.slice(0, 12);

        for (const p of c.Ports || []) {
          if (p.PublicPort) {
            const conflict = seen.has(p.PublicPort) && seen.get(p.PublicPort) !== service;
            portMap.push({
              container: c.Id.slice(0, 12),
              service,
              port: p.PublicPort,
              protocol: p.Type || "tcp",
              status: conflict ? "conflict" : "open",
              stack,
            });
            seen.set(p.PublicPort, service);
          }
        }
      }

      portMap.sort((a, b) => a.port - b.port);
      setPorts(portMap);
    } catch (err) {
      ddClient.desktopUI.toast.error("Failed to fetch ports");
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchPorts();
    const interval = setInterval(fetchPorts, 10000);
    return () => clearInterval(interval);
  }, []);

  const openPort = (port: number) => {
    ddClient.host.openUrl(`http://localhost:${port}`);
  };

  if (loading) return <div style={{ padding: 24 }}>Scanning ports...</div>;

  const conflicts = ports.filter(p => p.status === "conflict");

  return (
    <div style={{ padding: 24, fontFamily: "system-ui, sans-serif" }}>
      <h1 style={{ fontSize: 24, marginBottom: 8 }}>Port Dashboard</h1>
      <p style={{ color: "#6b7280", marginBottom: 16 }}>
        {ports.length} ports exposed across {new Set(ports.map(p => p.stack)).size} stacks
      </p>

      {conflicts.length > 0 && (
        <div style={{ background: "#fef2f2", border: "1px solid #fecaca", borderRadius: 8, padding: 12, marginBottom: 16 }}>
          <strong style={{ color: "#dc2626" }}>Port Conflicts Detected:</strong>
          {conflicts.map(c => (
            <span key={c.port + c.service} style={{ marginLeft: 8, color: "#dc2626" }}>
              :{c.port} ({c.service})
            </span>
          ))}
        </div>
      )}

      <table style={{ width: "100%", borderCollapse: "collapse", fontSize: 14 }}>
        <thead>
          <tr style={{ textAlign: "left", borderBottom: "2px solid #e5e7eb" }}>
            <th style={{ padding: 8 }}>Port</th>
            <th style={{ padding: 8 }}>Service</th>
            <th style={{ padding: 8 }}>Stack</th>
            <th style={{ padding: 8 }}>Protocol</th>
            <th style={{ padding: 8 }}>Status</th>
            <th style={{ padding: 8 }}>Action</th>
          </tr>
        </thead>
        <tbody>
          {ports.map((p, i) => (
            <tr key={i} style={{ borderBottom: "1px solid #f3f4f6" }}>
              <td style={{ padding: 8, fontFamily: "monospace" }}>:{p.port}</td>
              <td style={{ padding: 8 }}>{p.service}</td>
              <td style={{ padding: 8, color: "#6b7280" }}>{p.stack}</td>
              <td style={{ padding: 8 }}>{p.protocol}</td>
              <td style={{ padding: 8 }}>
                <span style={{
                  color: p.status === "open" ? "#22c55e" : "#ef4444",
                  fontWeight: 600,
                }}>
                  {p.status}
                </span>
              </td>
              <td style={{ padding: 8 }}>
                <button onClick={() => openPort(p.port)} style={{ cursor: "pointer" }}>
                  Open
                </button>
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}

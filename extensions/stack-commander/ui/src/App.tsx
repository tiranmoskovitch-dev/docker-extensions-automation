import React, { useEffect, useState } from "react";
import { createDockerDesktopClient } from "@docker/extension-api-client";

const ddClient = createDockerDesktopClient();

interface Service {
  name: string;
  status: string;
  ports: number[];
  health: string;
}

interface Stack {
  name: string;
  status: string;
  services: Service[];
  path: string;
}

export function App() {
  const [stacks, setStacks] = useState<Stack[]>([]);
  const [loading, setLoading] = useState(true);

  const fetchStacks = async () => {
    try {
      const response = await ddClient.extension.vm?.service?.get("/api/stacks");
      setStacks(response as Stack[]);
    } catch (err) {
      ddClient.desktopUI.toast.error("Failed to fetch stacks");
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchStacks();
    const interval = setInterval(fetchStacks, 5000);
    return () => clearInterval(interval);
  }, []);

  const handleAction = async (stackName: string, action: string) => {
    try {
      await ddClient.extension.vm?.service?.post(`/api/stacks/${action}`, {
        name: stackName,
      });
      ddClient.desktopUI.toast.success(`Stack ${stackName}: ${action} initiated`);
      fetchStacks();
    } catch (err) {
      ddClient.desktopUI.toast.error(`Failed to ${action} stack ${stackName}`);
    }
  };

  const statusColor = (status: string) => {
    switch (status) {
      case "running": return "#22c55e";
      case "partial": return "#f59e0b";
      case "stopped": return "#ef4444";
      default: return "#6b7280";
    }
  };

  if (loading) {
    return <div style={{ padding: 24 }}>Loading stacks...</div>;
  }

  return (
    <div style={{ padding: 24, fontFamily: "system-ui, sans-serif" }}>
      <h1 style={{ fontSize: 24, marginBottom: 16 }}>Stack Commander</h1>
      <p style={{ color: "#6b7280", marginBottom: 24 }}>
        Manage all your Docker Compose stacks from one place.
      </p>

      <div style={{ display: "flex", gap: 8, marginBottom: 24 }}>
        <button onClick={() => stacks.forEach(s => handleAction(s.name, "start"))}>
          Start All
        </button>
        <button onClick={() => stacks.forEach(s => handleAction(s.name, "stop"))}>
          Stop All
        </button>
        <button onClick={fetchStacks}>Refresh</button>
      </div>

      {stacks.length === 0 ? (
        <p>No Compose stacks found. Start a stack with `docker compose up -d`.</p>
      ) : (
        <div style={{ display: "grid", gap: 16 }}>
          {stacks.map((stack) => (
            <div
              key={stack.name}
              style={{
                border: "1px solid #e5e7eb",
                borderRadius: 8,
                padding: 16,
                borderLeft: `4px solid ${statusColor(stack.status)}`,
              }}
            >
              <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center" }}>
                <div>
                  <h2 style={{ fontSize: 18, margin: 0 }}>{stack.name}</h2>
                  <span style={{ color: statusColor(stack.status), fontSize: 14 }}>
                    {stack.status}
                  </span>
                  {stack.path && (
                    <span style={{ color: "#9ca3af", fontSize: 12, marginLeft: 12 }}>
                      {stack.path}
                    </span>
                  )}
                </div>
                <div style={{ display: "flex", gap: 8 }}>
                  <button onClick={() => handleAction(stack.name, "start")}>Start</button>
                  <button onClick={() => handleAction(stack.name, "stop")}>Stop</button>
                  <button onClick={() => handleAction(stack.name, "restart")}>Restart</button>
                </div>
              </div>

              <div style={{ marginTop: 12 }}>
                <table style={{ width: "100%", fontSize: 14 }}>
                  <thead>
                    <tr style={{ textAlign: "left", color: "#6b7280" }}>
                      <th>Service</th>
                      <th>Status</th>
                      <th>Ports</th>
                    </tr>
                  </thead>
                  <tbody>
                    {stack.services.map((svc) => (
                      <tr key={svc.name}>
                        <td>{svc.name}</td>
                        <td style={{ color: statusColor(svc.status === "running" ? "running" : "stopped") }}>
                          {svc.status}
                        </td>
                        <td>{svc.ports?.join(", ") || "--"}</td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}

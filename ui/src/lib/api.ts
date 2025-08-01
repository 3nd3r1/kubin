// API client for Kubin server communication

export interface Snapshot {
  id: string;
  metadata: {
    timestamp: string;
    cluster: {
      name: string;
      version: string;
      context: string;
    };
    resources: {
      pods: number;
      logs: number;
    };
    filters: {
      namespaces: string[];
      labelSelectors: string[];
    };
  };
  files: {
    [key: string]: string;
  };
}

export interface Pod {
  kind: string;
  namespace: string;
  name: string;
  data: any;
  logs: Array<{
    container: string;
    logs: string;
  }>;
}

class KubinAPI {
  private baseUrl: string;

  constructor(baseUrl: string = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080') {
    this.baseUrl = baseUrl;
  }

  // Get snapshot metadata
  async getSnapshot(id: string): Promise<Snapshot> {
    const response = await fetch(`${this.baseUrl}/api/v1/snapshots/${id}`);
    if (!response.ok) {
      throw new Error(`Failed to fetch snapshot: ${response.statusText}`);
    }
    return response.json();
  }

  // Get resources for a snapshot
  async getSnapshotResources(id: string): Promise<Pod[]> {
    const response = await fetch(`${this.baseUrl}/internal/api/v1/snapshots/${id}/resources`);
    if (!response.ok) {
      throw new Error(`Failed to fetch resources: ${response.statusText}`);
    }
    return response.json();
  }

  // Get pods for a snapshot
  async getSnapshotPods(id: string): Promise<Pod[]> {
    const response = await fetch(`${this.baseUrl}/internal/api/v1/snapshots/${id}/pods`);
    if (!response.ok) {
      throw new Error(`Failed to fetch pods: ${response.statusText}`);
    }
    return response.json();
  }

  // Get logs for a specific pod
  async getPodLogs(id: string, namespace: string, podName: string): Promise<string> {
    const response = await fetch(`${this.baseUrl}/internal/api/v1/snapshots/${id}/logs?namespace=${namespace}&pod=${podName}`);
    if (!response.ok) {
      throw new Error(`Failed to fetch logs: ${response.statusText}`);
    }
    return response.text();
  }

  // Get namespaces for a snapshot
  async getSnapshotNamespaces(id: string): Promise<string[]> {
    const response = await fetch(`${this.baseUrl}/internal/api/v1/snapshots/${id}/namespaces`);
    if (!response.ok) {
      throw new Error(`Failed to fetch namespaces: ${response.statusText}`);
    }
    return response.json();
  }
}

// Export a singleton instance
export const kubinAPI = new KubinAPI(); 
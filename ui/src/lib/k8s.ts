// The parts of Kubernetes objects the UI reads. Snapshots hold whatever the
// API server returned, so every field is optional.

export type ObjectMeta = {
    name?: string;
    uid?: string;
    namespace?: string;
    creationTimestamp?: string;
    deletionTimestamp?: string;
    labels?: Record<string, string>;
    annotations?: Record<string, string>;
    ownerReferences?: {
        apiVersion?: string;
        kind?: string;
        name?: string;
        uid?: string;
        controller?: boolean;
    }[];
};

export type ContainerState = {
    running?: { startedAt?: string };
    waiting?: { reason?: string; message?: string };
    terminated?: {
        reason?: string;
        message?: string;
        exitCode?: number;
        signal?: number;
        startedAt?: string;
        finishedAt?: string;
    };
};

export type ContainerStatus = {
    name: string;
    image?: string;
    ready?: boolean;
    started?: boolean;
    restartCount?: number;
    state?: ContainerState;
    lastState?: ContainerState;
};

export type Container = {
    name: string;
    image?: string;
    envFrom?: { configMapRef?: { name?: string } }[];
    env?: { valueFrom?: { configMapKeyRef?: { name?: string } } }[];
};

export type Pod = {
    apiVersion?: string;
    kind?: string;
    metadata?: ObjectMeta;
    spec?: {
        nodeName?: string;
        containers?: Container[];
        initContainers?: (Container & { restartPolicy?: string })[];
        volumes?: {
            configMap?: { name?: string };
            projected?: { sources?: { configMap?: { name?: string } }[] };
        }[];
    };
    status?: {
        phase?: string;
        reason?: string;
        message?: string;
        podIP?: string;
        conditions?: {
            type: string;
            status: string;
            reason?: string;
            message?: string;
        }[];
        containerStatuses?: ContainerStatus[];
        initContainerStatuses?: ContainerStatus[];
    };
};

export type Event = {
    metadata?: ObjectMeta;
    involvedObject?: {
        kind?: string;
        namespace?: string;
        name?: string;
        uid?: string;
    };
    type?: string;
    reason?: string;
    message?: string;
    count?: number;
    firstTimestamp?: string | null;
    lastTimestamp?: string | null;
    eventTime?: string | null;
    series?: { count?: number; lastObservedTime?: string };
    source?: { component?: string };
    reportingComponent?: string;
};

type Condition = { type?: string; status?: string };
type Quantities = Record<string, string>;
type PortInfo = {
    port?: number;
    nodePort?: number;
    protocol?: string;
    name?: string;
};
type LoadBalancerStatus = {
    ingress?: { ip?: string; hostname?: string }[];
};

/**
 * One loose shape for the kinds columns.ts reads. Fields belong to different
 * kinds; each column reads only its own.
 */
export type AnyObject = {
    kind?: string;
    metadata?: ObjectMeta;
    // ConfigMap
    data?: Record<string, string>;
    binaryData?: Record<string, string>;
    // EndpointSlice
    addressType?: string;
    ports?: { port?: number; protocol?: string }[];
    endpoints?: { addresses?: string[] }[];
    // Event
    involvedObject?: Event["involvedObject"];
    regarding?: Event["involvedObject"];
    type?: string;
    reason?: string;
    message?: string;
    note?: string;
    firstTimestamp?: string | null;
    lastTimestamp?: string | null;
    eventTime?: string | null;
    series?: { count?: number; lastObservedTime?: string };
    spec?: {
        replicas?: number;
        suspend?: boolean;
        completions?: number;
        schedule?: string;
        timeZone?: string;
        unschedulable?: boolean;
        type?: string;
        clusterIP?: string;
        externalIPs?: string[];
        externalName?: string;
        ports?: PortInfo[];
        ingressClassName?: string;
        rules?: { host?: string }[];
        tls?: unknown[];
        volumeName?: string;
        storageClassName?: string;
        accessModes?: string[];
        capacity?: Quantities;
        persistentVolumeReclaimPolicy?: string;
        claimRef?: { namespace?: string; name?: string };
        selector?: Record<string, string>;
        template?: { spec?: { nodeSelector?: Record<string, string> } };
    };
    status?: {
        phase?: string;
        replicas?: number;
        readyReplicas?: number;
        updatedReplicas?: number;
        availableReplicas?: number;
        desiredNumberScheduled?: number;
        currentNumberScheduled?: number;
        numberReady?: number;
        updatedNumberScheduled?: number;
        numberAvailable?: number;
        succeeded?: number;
        startTime?: string;
        completionTime?: string;
        active?: number;
        lastScheduleTime?: string;
        conditions?: Condition[];
        loadBalancer?: LoadBalancerStatus;
        capacity?: Quantities;
        accessModes?: string[];
        nodeInfo?: { kubeletVersion?: string };
    };
};

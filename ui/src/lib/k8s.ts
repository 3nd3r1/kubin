// The parts of Kubernetes objects the UI reads. Snapshots hold whatever the
// API server returned, so every field is optional.

export type ObjectMeta = {
    name?: string;
    uid?: string;
    namespace?: string;
    creationTimestamp?: string;
    deletionTimestamp?: string;
    labels?: Record<string, string>;
    ownerReferences?: { kind?: string; name?: string }[];
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

export type Container = { name: string; image?: string };

export type Pod = {
    apiVersion?: string;
    kind?: string;
    metadata?: ObjectMeta;
    spec?: {
        nodeName?: string;
        containers?: Container[];
        initContainers?: (Container & { restartPolicy?: string })[];
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

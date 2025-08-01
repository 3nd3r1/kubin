import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
    Card,
    CardContent,
    CardDescription,
    CardHeader,
    CardTitle,
} from "@/components/ui/card";

export default function Home() {
    return (
        <div className="min-h-screen py-12">
            <div className="max-w-4xl mx-auto">
                {/* Header */}
                <div className="text-center mb-12">
                    <h1 className="text-4xl font-bold mb-4">
                        Kubin UI
                    </h1>
                    <p className="text-xl">
                        Lens-like interface for viewing Kubernetes cluster
                        snapshots
                    </p>
                    <div className="flex justify-center gap-2 mt-4">
                        <Badge variant="secondary">Next.js</Badge>
                        <Badge variant="secondary">TypeScript</Badge>
                        <Badge variant="secondary">Tailwind CSS</Badge>
                        <Badge variant="secondary">shadcn/ui</Badge>
                    </div>
                </div>

                {/* Main Content */}
                <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
                    {/* Snapshot Card */}
                    <Card>
                        <CardHeader>
                            <CardTitle className="flex items-center gap-2">
                                📸 Snapshot Viewer
                            </CardTitle>
                            <CardDescription>
                                View cluster state as it was at snapshot time
                            </CardDescription>
                        </CardHeader>
                        <CardContent>
                            <p className="text-sm mb-4">
                                Explore your Kubernetes cluster snapshots with a
                                familiar Lens-like interface.
                            </p>
                            <Button className="w-full">View Snapshots</Button>
                        </CardContent>
                    </Card>

                    {/* Resource Browser Card */}
                    <Card>
                        <CardHeader>
                            <CardTitle className="flex items-center gap-2">
                                🗂️ Resource Browser
                            </CardTitle>
                            <CardDescription>
                                Navigate through namespaces and resources
                            </CardDescription>
                        </CardHeader>
                        <CardContent>
                            <p className="text-sm mb-4">
                                Browse pods, services, and other Kubernetes
                                resources in a hierarchical view.
                            </p>
                            <Button variant="outline" className="w-full">
                                Browse Resources
                            </Button>
                        </CardContent>
                    </Card>

                    {/* Pod Details Card */}
                    <Card>
                        <CardHeader>
                            <CardTitle className="flex items-center gap-2">
                                📋 Pod Details
                            </CardTitle>
                            <CardDescription>
                                View pod information and logs
                            </CardDescription>
                        </CardHeader>
                        <CardContent>
                            <p className="text-sm mb-4">
                                Get detailed information about pods including
                                their logs and status.
                            </p>
                            <Button variant="outline" className="w-full">
                                View Pods
                            </Button>
                        </CardContent>
                    </Card>
                </div>

                {/* Footer */}
                <div className="text-center mt-12">
                    <p>Ready to explore your Kubernetes snapshots? 🚀</p>
                </div>
            </div>
        </div>
    );
}

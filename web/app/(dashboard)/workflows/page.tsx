import React from "react";
import { Metadata } from "next";
import { WorkflowBuilder } from "@/components/recovery/builder/WorkflowBuilder";

export const metadata: Metadata = {
  title: "Workflow Builder | Payment Watchdog",
  description: "Design and manage automated recovery workflows",
};

export default function WorkflowsPage() {
  return (
    <div className="flex flex-col gap-6 p-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold tracking-tight">
            Workflow Builder
          </h1>
          <p className="text-muted-foreground mt-1">
            Design dynamic recovery strategies using the visual workflow editor.
          </p>
        </div>
      </div>

      <WorkflowBuilder
        onSave={(steps) => {
          console.log("Saved steps:", steps);
          // In a real app, send to API:
          // fetch('/api/workflows', { method: 'POST', body: JSON.stringify(steps) })
        }}
      />
    </div>
  );
}

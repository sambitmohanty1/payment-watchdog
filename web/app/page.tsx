import React from 'react';

export default function HomePage() {
  return (
    <div className="min-h-screen bg-gray-50">
      <div className="container mx-auto px-4 py-8">
        <div className="text-center">
          <h1 className="text-4xl font-bold text-primary-600 mb-4">
            ⚡ Payment Watchdog
          </h1>
          <p className="text-xl text-gray-600 mb-8">
            AI-powered SaaS payment failure intelligence platform
          </p>

          <div className="grid grid-cols-1 md:grid-cols-3 gap-6 max-w-4xl mx-auto">
            <div className="bg-white rounded-lg shadow-lg p-6 border-t-4 border-success-500">
              <div className="text-3xl mb-4">✅</div>
              <h3 className="text-lg font-semibold mb-2">API Status</h3>
              <p className="text-gray-600">Backend services running</p>
            </div>

            <div className="bg-white rounded-lg shadow-lg p-6 border-t-4 border-primary-500">
              <div className="text-3xl mb-4">🤖</div>
              <h3 className="text-lg font-semibold mb-2">AI Analytics</h3>
              <p className="text-gray-600">Pattern detection active</p>
            </div>

            <div className="bg-white rounded-lg shadow-lg p-6 border-t-4 border-warning-500">
              <div className="text-3xl mb-4">🚀</div>
              <h3 className="text-lg font-semibold mb-2">Worker Service</h3>
              <p className="text-gray-600">Event processing running</p>
            </div>
          </div>

          <div className="mt-12 bg-white rounded-lg shadow-lg p-6 max-w-2xl mx-auto">
            <h2 className="text-2xl font-semibold mb-4">System Status</h2>
            <div className="space-y-2 text-left">
              <div className="flex justify-between">
                <span>API Server:</span>
                <span className="text-success-600">✅ Running on port 8080</span>
              </div>
              <div className="flex justify-between">
                <span>Web Interface:</span>
                <span className="text-success-600">✅ Running on port 4896</span>
              </div>
              <div className="flex justify-between">
                <span>Database:</span>
                <span className="text-success-600">✅ PostgreSQL healthy</span>
              </div>
              <div className="flex justify-between">
                <span>Redis:</span>
                <span className="text-success-600">✅ Cache & Queue active</span>
              </div>
              <div className="flex justify-between">
                <span>Worker:</span>
                <span className="text-success-600">✅ Event processing active</span>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}

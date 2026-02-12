/** @type {import('next').NextConfig} */
const nextConfig = {
  // Removed standalone output to enable API routes
  // output: 'standalone',
  experimental: {
    appDir: true,
    serverComponentsExternalPackages: [],
  },
  
  // HTTPS configuration removed - using custom server instead
  
  // Handle proxy issues in development
  async headers() {
    return [
      {
        source: '/api/backend/:path*',
        headers: [
          {
            key: 'Cache-Control',
            value: 'no-cache, no-store, must-revalidate',
          },
          {
            key: 'Pragma',
            value: 'no-cache',
          },
          {
            key: 'Expires',
            value: '0',
          },
          {
            key: 'X-Proxy-Bypass',
            value: 'localhost,127.0.0.1,::1',
          },
        ],
      },
    ];
  },
  
  // Add proxy bypass for development
  async rewrites() {
    return [
      {
        source: '/api/backend/:path*',
        destination: 'http://localhost:8080/api/:path*',
      },
    ];
  },
  
  // Environment-specific settings
  env: {
    CUSTOM_KEY: process.env.CUSTOM_KEY || '',
  },
  
  // Handle proxy issues in webpack
  webpack: (config, { dev, isServer }) => {
    if (dev && !isServer) {
      // Add proxy bypass for development
      config.resolve.fallback = {
        ...config.resolve.fallback,
        net: false,
        tls: false,
      };
    }
    return config;
  },
};

module.exports = nextConfig;

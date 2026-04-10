/** @type {import('next').NextConfig} */
const nextConfig = {
  // Enable standalone output for Docker builds
  output: 'standalone',
  experimental: {
    // appDir is now default in Next.js 14
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
  
  // Add proxy bypass for development and production internal routing
  async rewrites() {
    return [
      {
        source: '/api/backend/:path*',
        destination: process.env.INTERNAL_API_URL 
          ? `${process.env.INTERNAL_API_URL}/api/:path*`
          : 'http://localhost:8080/api/:path*',
      },
      {
        source: '/api/system/:path*',
        destination: process.env.INTERNAL_API_URL 
          ? `${process.env.INTERNAL_API_URL}/api/:path*`
          : 'http://localhost:8085/:path*',
      },
    ];
  },
  
  // Environment-specific settings
  env: {
    CUSTOM_KEY: process.env.CUSTOM_KEY || '',
  },
  
  // Handle proxy issues in webpack - DISABLED FOR NEXT.JS 16
  // webpack: (config, { dev, isServer }) => {
  //   if (dev && !isServer) {
  //     // Add proxy bypass for development
  //     config.resolve.fallback = {
  //       ...config.resolve.fallback,
  //       net: false,
  //       tls: false,
  //     };
  //   }
  //   return config;
  // },
};

module.exports = nextConfig;

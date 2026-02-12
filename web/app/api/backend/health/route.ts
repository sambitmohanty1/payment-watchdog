import { NextRequest, NextResponse } from 'next/server';

// Use localhost for local development (port-forwarded), internal service for production
const BACKEND_URL = process.env.NODE_ENV === 'development'
  ? 'http://localhost:8080'
  : 'http://lexure-intelligence-mvp.lexure-mvp.svc.cluster.local:8085';

export async function GET(request: NextRequest) {
  try {
    const url = new URL(request.url);
    const queryString = url.search;
    
    // Use the correct backend endpoint path
    const backendUrl = `${BACKEND_URL}/health${queryString}`;
    
    // Add proxy bypass headers and timeout
    const controller = new AbortController();
    const timeoutId = setTimeout(() => controller.abort(), 10000); // 10 second timeout
    
    const response = await fetch(backendUrl, {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json',
        'X-Forwarded-For': request.ip || 'unknown',
        'User-Agent': request.headers.get('user-agent') || 'unknown',
        // Note: 'Connection' header is blocked by browsers for security reasons
      },
      signal: controller.signal,
    });

    clearTimeout(timeoutId);

    if (!response.ok) {
      console.error('Backend health check failed:', response.status, response.statusText);
      return NextResponse.json(
        { error: 'Backend request failed', status: response.status },
        { status: response.status }
      );
    }

    const data = await response.json();
    
    // Add security headers to response
    const responseHeaders = new Headers();
    responseHeaders.set('X-Content-Type-Options', 'nosniff');
    responseHeaders.set('X-Frame-Options', 'DENY');
    responseHeaders.set('X-XSS-Protection', '1; mode=block');
    responseHeaders.set('Cache-Control', 'no-cache, no-store, must-revalidate');
    responseHeaders.set('Pragma', 'no-cache');
    responseHeaders.set('Expires', '0');
    
    return NextResponse.json(data, { headers: responseHeaders });
  } catch (error) {
    console.error('Health check error:', error);
    
    // Handle specific proxy-related errors
    if (error instanceof Error) {
      if (error.name === 'AbortError') {
        return NextResponse.json(
          { error: 'Backend request timeout - check proxy configuration' },
          { status: 504 }
        );
      }
      if (error.message.includes('ECONNREFUSED') || error.message.includes('ENOTFOUND')) {
        return NextResponse.json(
          { error: 'Backend service unreachable - check port forwarding and proxy settings' },
          { status: 503 }
        );
      }
    }
    
    return NextResponse.json(
      { error: 'Internal server error - check proxy configuration' },
      { status: 500 }
    );
  }
}

import { NextRequest, NextResponse } from 'next/server';

// Use localhost for local development (port-forwarded), internal service for production
const BACKEND_URL = process.env.NODE_ENV === 'development' 
  ? 'http://localhost:8085'
  : 'http://lexure-intelligence-mvp.lexure-mvp.svc.cluster.local:8085';

// Security middleware - validate company_id and other parameters
function validateRequest(request: NextRequest): { isValid: boolean; error?: string } {
  const url = new URL(request.url);
  const companyId = url.searchParams.get('company_id');
  
  // Validate company ID format (UUID)
  if (!companyId || !/^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i.test(companyId)) {
    return { isValid: false, error: 'Invalid company ID format' };
  }
  
  return { isValid: true };
}

export async function GET(request: NextRequest) {
  try {
    // Security validation
    const validation = validateRequest(request);
    if (!validation.isValid) {
      return NextResponse.json(
        { error: validation.error },
        { status: 400 }
      );
    }

    const url = new URL(request.url);
    const queryString = url.search;
    
    const backendUrl = `${BACKEND_URL}/api/v1/failures${queryString}`;
    
    const response = await fetch(backendUrl, {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json',
        'X-Forwarded-For': request.ip || 'unknown',
        'User-Agent': request.headers.get('user-agent') || 'unknown',
      },
    });

    if (!response.ok) {
      return NextResponse.json(
        { error: 'Backend request failed' },
        { status: response.status }
      );
    }

    const data = await response.json();
    
    // Add security headers to response
    const responseHeaders = new Headers();
    responseHeaders.set('X-Content-Type-Options', 'nosniff');
    responseHeaders.set('X-Frame-Options', 'DENY');
    responseHeaders.set('X-XSS-Protection', '1; mode=block');
    
    return NextResponse.json(data, { headers: responseHeaders });
  } catch (error) {
    console.error('Proxy error:', error);
    return NextResponse.json(
      { error: 'Internal server error' },
      { status: 500 }
    );
  }
}

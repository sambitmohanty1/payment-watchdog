import { NextRequest, NextResponse } from 'next/server';

const API_BASE_URL = 'http://lexure-intelligence-mvp.lexure-mvp.svc.cluster.local:8085';

export async function GET(request: NextRequest) {
  try {
    console.log('Xero test API - API_BASE_URL:', API_BASE_URL);
    const url = `${API_BASE_URL}/api/v1/xero/test`;
    console.log('Xero test API - Full URL:', url);
    
    const response = await fetch(url, {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json',
      },
    });
    
    console.log('Xero test API - Response status:', response.status);
    const data = await response.json();
    console.log('Xero test API - Response data:', data);
    
    return NextResponse.json(data, {
      status: response.status,
      headers: {
        'Access-Control-Allow-Origin': '*',
        'Access-Control-Allow-Methods': 'GET, POST, PUT, DELETE, OPTIONS',
        'Access-Control-Allow-Headers': 'Content-Type, Authorization, company_id',
      },
    });
  } catch (error) {
    console.error('Xero test API error:', error);
    return NextResponse.json(
      { error: 'Failed to connect to Xero API', details: error instanceof Error ? error.message : 'Unknown error' },
      { status: 500 }
    );
  }
}

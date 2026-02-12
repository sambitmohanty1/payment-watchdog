import { d as devtools } from './devtools-PtxSnd7z.cjs';
import 'react';
import '@tanstack/query-devtools';
import '@tanstack/react-query';

declare const ReactQueryDevtools: (typeof devtools)['ReactQueryDevtools'];

export { ReactQueryDevtools };

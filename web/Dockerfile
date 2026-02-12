# Multi-stage build for Next.js application
FROM node:18-alpine AS base

# Install dependencies only when needed
FROM base AS deps
# Check https://github.com/nodejs/docker-node/tree/b4117f9333da4138b03a546ec926ef50a31506c3#nodealpine to understand why libc6-compat might be needed.
RUN apk add --no-cache libc6-compat
WORKDIR /app

# Install dependencies based on the preferred package manager
COPY package.json package-lock.json* ./
RUN npm ci

# Rebuild the source code only when needed
FROM base AS builder
WORKDIR /app
COPY --from=deps /app/node_modules ./node_modules
COPY . .

# Set proxy bypass environment variables
ENV NO_PROXY=localhost,127.0.0.1,::1
ENV no_proxy=localhost,127.0.0.1,::1
ENV NEXT_PUBLIC_DEV_PROXY_BYPASS=true
ENV NEXT_PUBLIC_LOCALHOST_BYPASS=true

# Build the application
RUN npm run build

# Production image, copy all the files and run next
FROM base AS runner
WORKDIR /app

ENV NODE_ENV=production
ENV PORT=4896

# Set proxy bypass environment variables for production
ENV NO_PROXY=localhost,127.0.0.1,::1,lexure-intelligence-mvp.lexure-mvp.svc.cluster.local
ENV no_proxy=localhost,127.0.0.1,::1,lexure-intelligence-mvp.lexure-mvp.svc.cluster.local
ENV NEXT_PUBLIC_DEV_PROXY_BYPASS=true
ENV NEXT_PUBLIC_LOCALHOST_BYPASS=true

RUN addgroup --system --gid 1001 nodejs
RUN adduser --system --uid 1001 nextjs

# Copy the public folder
COPY --from=builder /app/public ./public

# Copy SSL certificates for HTTPS support (if they exist)
COPY --from=builder /app/certs ./certs

# Set the correct permission for prerender cache
RUN mkdir .next
RUN chown nextjs:nodejs .next

# Automatically leverage output traces to reduce image size
# https://nextjs.org/docs/advanced-features/output-file-tracing
COPY --from=builder --chown=nextjs:nodejs /app/.next ./.next
COPY --from=builder --chown=nextjs:nodejs /app/node_modules ./node_modules
COPY --from=builder --chown=nextjs:nodejs /app/package.json ./package.json

USER nextjs

EXPOSE 4896

ENV HOSTNAME="0.0.0.0"

CMD ["npm", "start"]

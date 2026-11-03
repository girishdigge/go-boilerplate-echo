import { z } from 'zod';

export const HealthCheckSchema = z.object({
  status: z.string(),
  response_time: z.string(),
  error: z.string().optional(),
});

export const HealthResponseSchema = z.object({
  status: z.enum(['healthy', 'unhealthy']),
  timestamp: z.iso.datetime({ offset: true }),
  environment: z.string(),
  checks: z.object({
    database: HealthCheckSchema,
    redis: HealthCheckSchema.optional(),
  }),
});

export type HealthCheck = z.infer<typeof HealthCheckSchema>;

export type HealthResponse = z.infer<typeof HealthResponseSchema>;

/**
 * @deprecated Use HealthResponseSchema instead.
 */
export const ZHealthResponse = HealthResponseSchema;

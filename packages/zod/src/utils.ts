import { z } from 'zod';

export type PaginatedResponse<T> = {
  data: T[];
  total: number;
  page: number;
  limit: number;
  totalPages: number;
};

export const paginatedResponseSchema = <T extends z.ZodType>(itemSchema: T) =>
  z.object({
    data: z.array(itemSchema),
    total: z.number(),
    page: z.number(),
    limit: z.number(),
    totalPages: z.number(),
  });

/**
 * @deprecated Use paginatedResponseSchema instead.
 */
export const schemaWithPagination = paginatedResponseSchema;

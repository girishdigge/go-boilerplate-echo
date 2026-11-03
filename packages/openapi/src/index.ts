import { createDocument } from 'zod-openapi';
import type { oas32 } from 'zod-openapi';
import { HealthCheckSchema, HealthResponseSchema } from '@boilerplate/zod';

import { paths } from './contracts/index.js';

export const OpenAPI: oas32.OpenAPIObject = createDocument({
  openapi: '3.1.0',

  info: {
    version: '1.0.0',
    title: 'Boilerplate REST API - Documentation',
    description: 'Boilerplate REST API - Documentation',
  },

  servers: [
    {
      url: 'http://localhost:8080',
      description: 'Local Server',
    },
  ],

  paths,

  components: {
    schemas: {
      HealthCheck: HealthCheckSchema,
      HealthResponse: HealthResponseSchema,
    },

    securitySchemes: {
      bearerAuth: {
        type: 'http',
        scheme: 'bearer',
        bearerFormat: 'JWT',
      },

      'x-service-token': {
        type: 'apiKey',
        name: 'x-service-token',
        in: 'header',
      },
    },
  },
});

export { getSecurityRequirements } from './utils.js';

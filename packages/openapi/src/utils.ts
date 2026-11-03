export type SecurityType = 'bearer' | 'service';

export type SecurityRequirements = Array<Record<string, string[]>>;

export const getSecurityRequirements = (
  securityType: SecurityType = 'bearer',
): SecurityRequirements => {
  switch (securityType) {
    case 'bearer':
      return [{ bearerAuth: [] }];

    case 'service':
      return [{ 'x-service-token': [] }];
  }
};

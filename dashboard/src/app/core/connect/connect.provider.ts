import { createTransport } from '@connectrpc/connect-web';
import { createClient } from '@connectrpc/connect';
import { environment } from '../../../environments/environment';

export function createConnectTransport() {
  return createTransport({
    baseUrl: environment.apiBaseUrl,
  });
}

import { createConnectTransport } from '@connectrpc/connect-web';
import { createClient } from '@connectrpc/connect';
import { environment } from '../../../environments/environment';

export function getConnectTransport() {
  return createConnectTransport({
    baseUrl: environment.apiBaseUrl,
  });
}

import { HttpInterceptorFn } from '@angular/common/http';

export const authInterceptor: HttpInterceptorFn = (req, next) => {
  // Since tokens are stored in HttpOnly cookies, we just need to ensure
  // that credentials are sent with every request to the API.
  const cloned = req.clone({
    withCredentials: true,
  });

  return next(cloned);
};

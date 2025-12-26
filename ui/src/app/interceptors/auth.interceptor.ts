import { HttpInterceptorFn } from '@angular/common/http';

export const authInterceptor: HttpInterceptorFn = (req, next) => {
  // Add withCredentials to send cookies with cross-origin requests
  const clonedRequest = req.clone({
    withCredentials: true
  });
  return next(clonedRequest);
};

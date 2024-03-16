const privateRoutes = [
    "/profile",
    "/lists",
  ]
  
  export function isRouteProtected(path: string) {
    return privateRoutes.some(allowedPath =>
      path === allowedPath || path.startsWith(allowedPath + '/')
    );
  }
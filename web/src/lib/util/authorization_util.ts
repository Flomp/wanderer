const privateRoutes = [
    "/profile",
    "/lists",
    "/trail",
  ]
  
  export function isRouteProtected(path: string) {
    return privateRoutes.some(allowedPath =>
      path === allowedPath || path.startsWith(allowedPath + '/')
    );
  }
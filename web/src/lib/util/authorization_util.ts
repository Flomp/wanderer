const privateRoutes = [
  "/settings",
  "/trail/edit/new",
  "/profile",
  "/lists/edit/new",
]

export function isRouteProtected(path: string) {
  return privateRoutes.some(allowedPath =>
    path === allowedPath || path.startsWith(allowedPath + '/')
  );
}
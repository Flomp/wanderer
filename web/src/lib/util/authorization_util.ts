const privateRoutes = [
  "/profile",
  "/lists",
  "/trail/edit/new",
]

export function isRouteProtected(path: string) {
  return privateRoutes.some(allowedPath =>
    path === allowedPath || path.startsWith(allowedPath + '/')
  );
}
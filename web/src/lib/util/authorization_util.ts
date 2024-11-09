const privateRoutes = [
  "/settings",
  "/lists",
  "/trail/edit/new",
  "/profile"
]

export function isRouteProtected(path: string) {
  return privateRoutes.some(allowedPath =>
    path === allowedPath || path.startsWith(allowedPath + '/')
  );
}
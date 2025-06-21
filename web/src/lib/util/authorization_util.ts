import { env } from "$env/dynamic/public";

const privateRoutes = [
  "/settings",
  "/trail/edit/new",
  "/profile",
  "/lists/edit/new",
]

const publicRoutes = [
  "/",
  "/login",
  "/api/v1/auth",
  "/api/v1/user",
  "/api/v1/category",
  "/register",
  "/auth"

]

export function isRouteProtected(path: string) {

  if (env.PUBLIC_PRIVATE_INSTANCE == "true") {
    return !publicRoutes.some(allowedPath =>
      path === allowedPath || path.startsWith(allowedPath + '/')
    );
  }

  return privateRoutes.some(allowedPath =>
    path === allowedPath || path.startsWith(allowedPath + '/')
  );
}
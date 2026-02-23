import { useTranslation } from "react-i18next";
import { Link, Outlet, useLocation } from "react-router-dom";
import { Search, Heart, Flame, LogIn, UserPlus, BarChart3, Bell, User, Menu, X } from "lucide-react";
import { useState } from "react";
import { useAuth } from "@/shared/contexts/AuthContext";
import { LanguageSwitcher } from "@/shared/components/LanguageSwitcher";
import { Button } from "@/components/ui/button";
import Footer from "@/shared/components/Footer";

const publicNav = [
  { path: "/missing", icon: Search, labelKey: "nav.missing" },
  { path: "/homeless", icon: Heart, labelKey: "nav.homeless" },
  { path: "/heatmap", icon: Flame, labelKey: "nav.heatmap" },
];

const authNav = [
  { path: "/dashboard", icon: BarChart3, labelKey: "nav.dashboard" },
  { path: "/notifications", icon: Bell, labelKey: "nav.notifications" },
  { path: "/profile", icon: User, labelKey: "nav.profile" },
];

export default function PublicLayout() {
  const { t } = useTranslation();
  const { user } = useAuth();
  const location = useLocation();
  const [mobileOpen, setMobileOpen] = useState(false);

  const navItems = user ? [...publicNav, ...authNav] : publicNav;

  return (
    <div className="flex min-h-screen flex-col">
      <a
        href="#main-content"
        className="sr-only focus:not-sr-only focus:absolute focus:z-50 focus:p-2 focus:bg-primary focus:text-primary-foreground"
      >
        {t("a11y.skipToContent")}
      </a>

      <header className="sticky top-0 z-40 border-b bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/60" role="banner">
        <div className="container mx-auto flex h-14 items-center justify-between px-4">
          <Link to="/" className="text-xl font-bold tracking-tight">
            Traceo
          </Link>

          <nav className="hidden md:flex items-center gap-1">
            {navItems.map((item) => {
              const active = location.pathname === item.path;
              return (
                <Link key={item.path} to={item.path}>
                  <Button
                    variant={active ? "secondary" : "ghost"}
                    size="sm"
                    className="gap-1.5"
                  >
                    <item.icon className="h-4 w-4" />
                    {t(item.labelKey)}
                  </Button>
                </Link>
              );
            })}
          </nav>

          <div className="flex items-center gap-2">
            <LanguageSwitcher />
            {!user ? (
              <div className="hidden md:flex items-center gap-2">
                <Link to="/login">
                  <Button variant="ghost" size="sm" className="gap-1.5">
                    <LogIn className="h-4 w-4" />
                    {t("nav.login")}
                  </Button>
                </Link>
                <Link to="/register">
                  <Button size="sm" className="gap-1.5">
                    <UserPlus className="h-4 w-4" />
                    {t("nav.register")}
                  </Button>
                </Link>
              </div>
            ) : (
              <Link to="/profile" className="hidden md:block">
                <Button variant="ghost" size="sm" className="gap-1.5">
                  <User className="h-4 w-4" />
                  {user.displayName || user.email?.split("@")[0]}
                </Button>
              </Link>
            )}
            <Button
              variant="ghost"
              size="icon"
              className="md:hidden"
              onClick={() => setMobileOpen(!mobileOpen)}
            >
              {mobileOpen ? <X className="h-5 w-5" /> : <Menu className="h-5 w-5" />}
            </Button>
          </div>
        </div>

        {mobileOpen && (
          <div className="md:hidden border-t bg-background p-4 space-y-1">
            {navItems.map((item) => {
              const active = location.pathname === item.path;
              return (
                <Link key={item.path} to={item.path} onClick={() => setMobileOpen(false)}>
                  <Button
                    variant={active ? "secondary" : "ghost"}
                    className="w-full justify-start gap-2"
                  >
                    <item.icon className="h-4 w-4" />
                    {t(item.labelKey)}
                  </Button>
                </Link>
              );
            })}
            {!user && (
              <>
                <Link to="/login" onClick={() => setMobileOpen(false)}>
                  <Button variant="ghost" className="w-full justify-start gap-2">
                    <LogIn className="h-4 w-4" />
                    {t("nav.login")}
                  </Button>
                </Link>
                <Link to="/register" onClick={() => setMobileOpen(false)}>
                  <Button className="w-full justify-start gap-2">
                    <UserPlus className="h-4 w-4" />
                    {t("nav.register")}
                  </Button>
                </Link>
              </>
            )}
          </div>
        )}
      </header>

      <main id="main-content" className="flex-1" role="main">
        <div className="container mx-auto px-4 py-6">
          <Outlet />
        </div>
      </main>

      <Footer />
    </div>
  );
}

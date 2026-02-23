import { useTranslation } from "react-i18next";
import { Link, Outlet, useLocation } from "react-router-dom";
import { Home, User, Key, LogOut, Search, BarChart3 } from "lucide-react";
import { useAuth } from "@/shared/contexts/AuthContext";
import { LanguageSwitcher } from "@/shared/components/LanguageSwitcher";
import { Button } from "@/components/ui/button";
import { Avatar, AvatarFallback } from "@/components/ui/avatar";
import { Separator } from "@/components/ui/separator";
import SearchBar from "@/shared/components/SearchBar";

const navItems = [
  { path: "/", icon: Home, labelKey: "nav.home" },
  { path: "/missing", icon: Search, labelKey: "nav.missing" },
  { path: "/dashboard", icon: BarChart3, labelKey: "nav.dashboard" },
  { path: "/profile", icon: User, labelKey: "nav.profile" },
  { path: "/password", icon: Key, labelKey: "nav.password" },
];

export default function AppLayout() {
  const { t } = useTranslation();
  const { user, signOut } = useAuth();
  const location = useLocation();

  const initials = user?.displayName
    ? user.displayName.slice(0, 2).toUpperCase()
    : user?.email?.slice(0, 2).toUpperCase() || "??";

  return (
    <div className="flex min-h-screen">
      <aside className="hidden w-64 flex-col border-r bg-muted/40 md:flex">
        <div className="flex h-14 items-center px-6 font-bold text-lg">
          Traceo
        </div>
        <Separator />
        <nav className="flex-1 space-y-1 p-4">
          {navItems.map((item) => {
            const active = location.pathname === item.path;
            return (
              <Link key={item.path} to={item.path}>
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
        </nav>
        <div className="p-4">
          <Button
            variant="ghost"
            className="w-full justify-start gap-2 text-muted-foreground"
            onClick={() => signOut()}
          >
            <LogOut className="h-4 w-4" />
            {t("nav.logout")}
          </Button>
        </div>
      </aside>

      <div className="flex flex-1 flex-col">
        <header className="flex h-14 items-center justify-between border-b px-6 gap-4">
          <h1 className="text-lg font-semibold md:hidden">Traceo</h1>
          <div className="hidden md:block flex-1 max-w-md">
            <SearchBar />
          </div>
          <div className="flex items-center gap-4 ml-auto">
            <LanguageSwitcher />
            <Avatar className="h-8 w-8">
              <AvatarFallback className="text-xs">{initials}</AvatarFallback>
            </Avatar>
          </div>
        </header>
        <main className="flex-1 p-6">
          <Outlet />
        </main>
      </div>
    </div>
  );
}

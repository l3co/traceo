import { useTranslation } from "react-i18next";
import { Link } from "react-router-dom";
import { SearchX } from "lucide-react";
import { Button } from "@/components/ui/button";
import Head from "@/shared/components/Head";

export default function NotFoundPage() {
  const { t } = useTranslation();

  return (
    <>
      <Head title="404" description={t("errors.notFoundDescription")} />
      <div className="flex flex-col items-center justify-center py-20 space-y-6 text-center">
        <SearchX className="h-20 w-20 text-muted-foreground" />
        <h1 className="text-4xl font-bold">404</h1>
        <p className="text-muted-foreground text-lg max-w-md">
          {t("errors.notFoundDescription")}
        </p>
        <Button asChild>
          <Link to="/">{t("errors.backHome")}</Link>
        </Button>
      </div>
    </>
  );
}

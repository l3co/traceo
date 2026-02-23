import { useTranslation } from "react-i18next";
import { Link } from "react-router-dom";
import { Search, Heart } from "lucide-react";
import Head from "@/shared/components/Head";
import { Card, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";

export default function ChoicePage() {
  const { t } = useTranslation();

  return (
    <>
      <Head title={t("choice.title")} />
      <div className="flex flex-col items-center justify-center min-h-[60vh] space-y-8">
        <h1 className="text-3xl font-bold text-center">{t("choice.title")}</h1>
        <p className="text-muted-foreground text-center max-w-md">{t("choice.subtitle")}</p>

        <div className="grid grid-cols-1 md:grid-cols-2 gap-6 w-full max-w-2xl">
          <Link to="/missing/new" className="group">
            <Card className="h-full transition-shadow hover:shadow-lg cursor-pointer border-2 hover:border-primary">
              <CardHeader className="flex flex-col items-center text-center space-y-4 py-8">
                <Search className="h-12 w-12 text-primary" />
                <CardTitle>{t("choice.registerMissing")}</CardTitle>
                <CardDescription>{t("choice.registerMissingDesc")}</CardDescription>
              </CardHeader>
            </Card>
          </Link>

          <Link to="/homeless/new" className="group">
            <Card className="h-full transition-shadow hover:shadow-lg cursor-pointer border-2 hover:border-primary">
              <CardHeader className="flex flex-col items-center text-center space-y-4 py-8">
                <Heart className="h-12 w-12 text-primary" />
                <CardTitle>{t("choice.wantToBeFound")}</CardTitle>
                <CardDescription>{t("choice.wantToBeFoundDesc")}</CardDescription>
              </CardHeader>
            </Card>
          </Link>
        </div>
      </div>
    </>
  );
}

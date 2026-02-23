import { useTranslation } from "react-i18next";
import Head from "@/shared/components/Head";

export default function TermsPage() {
  const { t } = useTranslation();

  return (
    <>
      <Head title={t("terms.title")} />
      <div className="max-w-2xl mx-auto prose prose-sm dark:prose-invert">
        <h1>{t("terms.title")}</h1>
        <p className="text-muted-foreground text-sm">{t("terms.lastUpdated")}</p>

        <h2>{t("terms.section1Title")}</h2>
        <p>{t("terms.section1Text")}</p>

        <h2>{t("terms.section2Title")}</h2>
        <p>{t("terms.section2Text")}</p>

        <h2>{t("terms.section3Title")}</h2>
        <p>{t("terms.section3Text")}</p>

        <h2>{t("terms.section4Title")}</h2>
        <p>{t("terms.section4Text")}</p>
      </div>
    </>
  );
}

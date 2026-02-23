import { useTranslation } from "react-i18next";
import Head from "@/shared/components/Head";

export default function PrivacyPage() {
  const { t } = useTranslation();

  return (
    <>
      <Head title={t("privacy.title")} />
      <div className="max-w-2xl mx-auto prose prose-sm dark:prose-invert">
        <h1>{t("privacy.title")}</h1>
        <p className="text-muted-foreground text-sm">{t("privacy.lastUpdated")}</p>

        <h2>{t("privacy.section1Title")}</h2>
        <p>{t("privacy.section1Text")}</p>

        <h2>{t("privacy.section2Title")}</h2>
        <p>{t("privacy.section2Text")}</p>

        <h2>{t("privacy.section3Title")}</h2>
        <p>{t("privacy.section3Text")}</p>

        <h2>{t("privacy.section4Title")}</h2>
        <p>{t("privacy.section4Text")}</p>

        <h2>{t("privacy.section5Title")}</h2>
        <p>{t("privacy.section5Text")}</p>
      </div>
    </>
  );
}

export const GENDER_OPTIONS = [
  { value: "male", labelPt: "Masculino", labelEn: "Male" },
  { value: "female", labelPt: "Feminino", labelEn: "Female" },
];

export const EYE_OPTIONS = [
  { value: "green", labelPt: "Verde", labelEn: "Green" },
  { value: "blue", labelPt: "Azul", labelEn: "Blue" },
  { value: "brown", labelPt: "Castanho", labelEn: "Brown" },
  { value: "black", labelPt: "Pretos", labelEn: "Black" },
  { value: "dark_brown", labelPt: "Castanho Escuro", labelEn: "Dark Brown" },
];

export const HAIR_OPTIONS = [
  { value: "black", labelPt: "Preto", labelEn: "Black" },
  { value: "brown", labelPt: "Castanho", labelEn: "Brown" },
  { value: "redhead", labelPt: "Ruivo", labelEn: "Redhead" },
  { value: "blond", labelPt: "Loiro", labelEn: "Blond" },
];

export const SKIN_OPTIONS = [
  { value: "white", labelPt: "Branca", labelEn: "White" },
  { value: "brown", labelPt: "Parda", labelEn: "Brown" },
  { value: "black", labelPt: "Negra", labelEn: "Black" },
  { value: "yellow", labelPt: "Amarela", labelEn: "Yellow" },
];

export const STATUS_OPTIONS = [
  { value: "disappeared", labelPt: "Desaparecido", labelEn: "Missing" },
  { value: "found", labelPt: "Encontrado", labelEn: "Found" },
];

export function getLabel(
  options: { value: string; labelPt: string; labelEn: string }[],
  value: string,
  lang: string
): string {
  const opt = options.find((o) => o.value === value);
  if (!opt) return value;
  return lang.startsWith("pt") ? opt.labelPt : opt.labelEn;
}

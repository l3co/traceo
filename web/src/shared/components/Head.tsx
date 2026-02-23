import { useEffect } from "react";

interface Props {
  title: string;
  description?: string;
}

export default function Head({ title, description }: Props) {
  useEffect(() => {
    document.title = `${title} | Traceo`;

    const metaDesc = document.querySelector('meta[name="description"]');
    if (description) {
      if (metaDesc) {
        metaDesc.setAttribute("content", description);
      } else {
        const meta = document.createElement("meta");
        meta.name = "description";
        meta.content = description;
        document.head.appendChild(meta);
      }
    }
  }, [title, description]);

  return null;
}

import DOMPurify from "dompurify";

const purifyConfig: DOMPurify.Config = {
    ALLOWED_TAGS: ["a", "b", "i", "em", "strong", "br", "p", "span"],
    ALLOWED_ATTR: ["href", "target", "rel", "class"],
};

export const sanitizeHtml = (dirty: string): string => {
    return DOMPurify.sanitize(dirty, purifyConfig);
};

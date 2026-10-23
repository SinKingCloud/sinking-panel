const fallbackCopyText = (value: string) => {
    if (typeof document === "undefined") {
        return false;
    }
    const activeElement = document.activeElement as HTMLElement | null;
    const textarea = document.createElement("textarea");
    try {
        textarea.value = value;
        textarea.readOnly = true;
        textarea.style.position = "fixed";
        textarea.style.top = "0";
        textarea.style.left = "0";
        textarea.style.opacity = "0";
        textarea.style.pointerEvents = "none";
        document.body.appendChild(textarea);
        textarea.select();
        textarea.setSelectionRange(0, value.length);
        return document.execCommand("copy");
    } catch {
        return false;
    } finally {
        textarea.remove();
        try {
            activeElement?.focus();
        } catch {
            // Ignore focus restoration failures after copying.
        }
    }
};

export const copyTextToClipboard = async (value: string) => {
    if (typeof window !== "undefined"
        && typeof navigator !== "undefined"
        && window.isSecureContext
        && navigator.clipboard?.writeText) {
        try {
            await navigator.clipboard.writeText(value);
            return true;
        } catch {
            // Fall back for browsers that expose the API but reject the write.
        }
    }
    return fallbackCopyText(value);
};

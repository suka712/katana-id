import { useRef, useState } from "react";

export function useFormDebounce<T extends Record<string, string>>(
  initialValues: T,
  delayMs: number = 3000
) {
  const lastSubmittedRef = useRef<T>(initialValues);
  const [isDebouncing, setIsDebouncing] = useState(false);

  const shouldSubmit = (currentValues: T): boolean => {
    const hasChanged = Object.keys(currentValues).some(
      (key) => currentValues[key] !== lastSubmittedRef.current[key]
    );
    if (!hasChanged || isDebouncing) return false;

    lastSubmittedRef.current = { ...currentValues };
    setIsDebouncing(true);
    setTimeout(() => setIsDebouncing(false), delayMs);
    return true;
  };

  return { isDebouncing, shouldSubmit };
}

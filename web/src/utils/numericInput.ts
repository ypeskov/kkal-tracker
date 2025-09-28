/**
 * Handles numeric input by replacing commas with dots and validating the format.
 * Allows decimal numbers with optional decimal point.
 *
 * @param value - The input string to process
 * @param setter - The function to call with the normalized value
 */
export const handleNumericInput = (value: string, setter: (value: string) => void): void => {
  // Replace all commas with dots for international decimal notation support
  const normalizedValue = value.replace(/,/g, '.');

  // Only allow valid number format: optional digits, optional single dot, optional more digits
  // Also allow empty string for clearing the input
  if (/^\d*\.?\d*$/.test(normalizedValue) || normalizedValue === '') {
    setter(normalizedValue);
  }
};
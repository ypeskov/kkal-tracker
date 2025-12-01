/**
 * Safely evaluates a mathematical expression
 * Supports: +, -, *, /
 * @param expr - The mathematical expression to evaluate
 * @returns The result of the calculation or null if invalid
 */
export const evaluateExpression = (expr: string): number | null => {
    try {
        // Remove all whitespace
        const cleanExpr = expr.replace(/\s/g, '');

        // Security: Only allow numbers, operators, dots, and parentheses
        if (!/^[\d+\-*/.()]+$/.test(cleanExpr)) {
            return null;
        }

        // Prevent empty expression
        if (!cleanExpr) {
            return null;
        }

        // Use Function constructor for safer evaluation than eval
        // eslint-disable-next-line no-new-func
        const result = new Function(`return ${cleanExpr}`)();

        // Check if result is a valid number
        if (typeof result === 'number' && !isNaN(result) && isFinite(result)) {
            return result;
        }

        return null;
    } catch (error) {
        return null;
    }
};

/**
 * Checks if a string contains mathematical operators
 * @param value - The string to check
 * @returns True if the string contains math operators
 */
export const isMathExpression = (value: string): boolean => {
    return /[+\-*/()]/.test(value);
};

/**
 * Formats a number to a reasonable precision for display
 * @param num - The number to format
 * @returns Formatted number as string
 */
export const formatCalculatorResult = (num: number): string => {
    // Round to 2 decimal places if needed
    const rounded = Math.round(num * 100) / 100;
    return rounded.toString();
};


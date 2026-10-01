/**
 * Safely evaluates a mathematical expression with a small recursive descent parser
 * (no eval / Function constructor, so the page works under a strict Content-Security-Policy)
 * Supports: +, -, *, /, parentheses, unary plus and minus
 * @param expr - The mathematical expression to evaluate
 * @returns The result of the calculation or null if invalid
 */
export const evaluateExpression = (expr: string): number | null => {
    // Remove all whitespace
    const cleanExpr = expr.replace(/\s/g, '');

    // Only allow numbers, operators, dots, and parentheses
    if (!cleanExpr || !/^[\d+\-*/.()]+$/.test(cleanExpr)) {
        return null;
    }

    let pos = 0;

    // expression = term (('+' | '-') term)*
    function parseExpression(): number {
        let value = parseTerm();
        while (cleanExpr[pos] === '+' || cleanExpr[pos] === '-') {
            const operator = cleanExpr[pos++];
            const right = parseTerm();
            value = operator === '+' ? value + right : value - right;
        }
        return value;
    }

    // term = factor (('*' | '/') factor)*
    function parseTerm(): number {
        let value = parseFactor();
        while (cleanExpr[pos] === '*' || cleanExpr[pos] === '/') {
            const operator = cleanExpr[pos++];
            const right = parseFactor();
            value = operator === '*' ? value * right : value / right;
        }
        return value;
    }

    // factor = ('+' | '-') factor | '(' expression ')' | number
    function parseFactor(): number {
        const char = cleanExpr[pos];
        if (char === '-') {
            pos++;
            return -parseFactor();
        }
        if (char === '+') {
            pos++;
            return parseFactor();
        }
        if (char === '(') {
            pos++;
            const value = parseExpression();
            if (cleanExpr[pos] !== ')') {
                throw new SyntaxError('Missing closing parenthesis');
            }
            pos++;
            return value;
        }
        return parseNumber();
    }

    // number = digits ['.' digits] | '.' digits
    function parseNumber(): number {
        const match = /^(\d+\.?\d*|\.\d+)/.exec(cleanExpr.slice(pos));
        if (!match) {
            throw new SyntaxError('Number expected');
        }
        pos += match[0].length;
        return parseFloat(match[0]);
    }

    try {
        const result = parseExpression();

        // The whole input must be consumed and the result must be a valid number
        if (pos === cleanExpr.length && !isNaN(result) && isFinite(result)) {
            return result;
        }

        return null;
    } catch {
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


import { useState, useRef, useEffect } from 'react';
import Calculator from './Calculator';
import { evaluateExpression, isMathExpression, formatCalculatorResult } from '@/utils/calculator';

interface CalculatorInputProps {
    value: string;
    onChange: (value: string) => void;
    className?: string;
    id?: string;
    required?: boolean;
    inputMode?: 'none' | 'text' | 'tel' | 'url' | 'email' | 'numeric' | 'decimal' | 'search';
    placeholder?: string;
    label?: string;
    disabled?: boolean;
}

export default function CalculatorInput({
    value,
    onChange,
    className = '',
    id,
    required,
    inputMode = 'decimal',
    placeholder,
    disabled
}: CalculatorInputProps) {
    const [isCalculatorExpanded, setIsCalculatorExpanded] = useState(false);
    const [calculatorPosition, setCalculatorPosition] = useState<{ top?: number; bottom?: number; left: number; right?: number }>({ top: 0, left: 0 });
    const inputRef = useRef<HTMLInputElement>(null);
    const containerRef = useRef<HTMLDivElement>(null);

    // Update calculator position when it's expanded
    useEffect(() => {
        if (isCalculatorExpanded && containerRef.current) {
            const rect = containerRef.current.getBoundingClientRect();
            const spaceBelow = window.innerHeight - rect.bottom;
            const spaceAbove = rect.top;

            // Decide whether to show calculator above or below the input
            if (spaceBelow > 320 || spaceBelow > spaceAbove) {
                // Show below
                setCalculatorPosition({
                    top: rect.height + 4,
                    left: 0
                });
            } else {
                // Show above
                setCalculatorPosition({
                    bottom: rect.height + 4,
                    left: 0
                });
            }
        }
    }, [isCalculatorExpanded]);

    const toggleCalculator = (e: React.MouseEvent) => {
        e.preventDefault();
        e.stopPropagation();
        const newExpandedState = !isCalculatorExpanded;
        setIsCalculatorExpanded(newExpandedState);

        // Focus input when calculator is opened
        if (newExpandedState && inputRef.current) {
            inputRef.current.focus();
        }
    };

    const handleInputChange = (newValue: string) => {
        // Use the existing handleNumericInput logic but also allow operators
        const normalizedValue = newValue.replace(/,/g, '.');

        // Allow numbers, operators, dots, and parentheses
        if (/^[\d+\-*/.()]*$/.test(normalizedValue) || normalizedValue === '') {
            onChange(normalizedValue);
        }
    };

    const handleCalculatorButtonClick = (buttonValue: string) => {
        if (buttonValue === 'C') {
            // Clear
            onChange('');
        } else if (buttonValue === '←') {
            // Backspace
            onChange(value.slice(0, -1));
        } else if (buttonValue === '=') {
            // Evaluate expression
            if (isMathExpression(value)) {
                const result = evaluateExpression(value);
                if (result !== null) {
                    onChange(formatCalculatorResult(result));
                }
            }
        } else {
            // Append button value to current input
            onChange(value + buttonValue);
        }

        // Keep focus on input after button click
        if (inputRef.current) {
            inputRef.current.focus();
        }
    };

    const handleKeyDown = (e: React.KeyboardEvent<HTMLInputElement>) => {
        // When Enter is pressed, try to evaluate the expression
        if (e.key === 'Enter' && isMathExpression(value)) {
            e.preventDefault();
            const result = evaluateExpression(value);
            if (result !== null) {
                onChange(formatCalculatorResult(result));
            }
        }
    };

    const handleBlur = () => {
        // When input loses focus, evaluate if it's a math expression
        if (isMathExpression(value)) {
            const result = evaluateExpression(value);
            if (result !== null) {
                onChange(formatCalculatorResult(result));
            }
        }
    };

    return (
        <div ref={containerRef} className="relative inline-block w-full">
            <div className="relative">
                <input
                    ref={inputRef}
                    type="text"
                    inputMode={inputMode}
                    id={id}
                    value={value}
                    onChange={(e) => handleInputChange(e.target.value)}
                    onKeyDown={handleKeyDown}
                    onBlur={handleBlur}
                    className={`pr-10 ${className}`}
                    required={required}
                    placeholder={placeholder}
                    disabled={disabled}
                />

                {/* Calculator icon button */}
                <button
                    type="button"
                    onClick={toggleCalculator}
                    onMouseDown={(e) => e.preventDefault()}
                    className={`
            absolute right-2 top-1/2 -translate-y-1/2 
            p-1 rounded transition-colors
            ${isCalculatorExpanded
                            ? 'text-blue-600 bg-blue-50'
                            : 'text-gray-400 hover:text-gray-600 hover:bg-gray-100'
                        }
          `}
                    aria-label="Toggle calculator"
                    disabled={disabled}
                >
                    <svg
                        xmlns="http://www.w3.org/2000/svg"
                        className="h-5 w-5"
                        viewBox="0 0 20 20"
                        fill="currentColor"
                    >
                        <path
                            fillRule="evenodd"
                            d="M6 2a2 2 0 00-2 2v12a2 2 0 002 2h8a2 2 0 002-2V4a2 2 0 00-2-2H6zm1 2a1 1 0 000 2h6a1 1 0 100-2H7zm6 7a1 1 0 011 1v3a1 1 0 11-2 0v-3a1 1 0 011-1zm-3 3a1 1 0 100 2h.01a1 1 0 100-2H10zm-4 1a1 1 0 011-1h.01a1 1 0 110 2H7a1 1 0 01-1-1zm1-4a1 1 0 100 2h.01a1 1 0 100-2H7zm2 1a1 1 0 011-1h.01a1 1 0 110 2H10a1 1 0 01-1-1zm4-4a1 1 0 100 2h.01a1 1 0 100-2H13zM9 9a1 1 0 011-1h.01a1 1 0 110 2H10a1 1 0 01-1-1zM7 8a1 1 0 000 2h.01a1 1 0 000-2H7z"
                            clipRule="evenodd"
                        />
                    </svg>
                </button>
            </div>

            {/* Calculator popup */}
            {isCalculatorExpanded && (
                <Calculator
                    isExpanded={isCalculatorExpanded}
                    onToggle={() => setIsCalculatorExpanded(false)}
                    onButtonClick={handleCalculatorButtonClick}
                    position={calculatorPosition}
                />
            )}
        </div>
    );
}


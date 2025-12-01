interface CalculatorProps {
    isExpanded: boolean;
    onToggle: () => void;
    onButtonClick: (value: string) => void;
    position?: { top?: number; bottom?: number; left: number; right?: number };
}

export default function Calculator({
    isExpanded,
    onToggle,
    onButtonClick,
    position = { top: 0, left: 0 }
}: CalculatorProps) {

    if (!isExpanded) {
        return null;
    }

    const buttons = [
        ['7', '8', '9', '/'],
        ['4', '5', '6', '*'],
        ['1', '2', '3', '-'],
        ['0', '.', '=', '+'],
        ['C', '←', '(', ')']
    ];

    const positionStyles: React.CSSProperties = {
        position: 'absolute',
        ...position
    };

    return (
        <div
            className="bg-white rounded-lg shadow-2xl border border-gray-300 p-3 z-50 animate-slideUp"
            style={positionStyles}
            onClick={(e) => e.stopPropagation()}
        >
            {/* Header with close button */}
            <div className="flex justify-between items-center mb-2 pb-2 border-b border-gray-200">
                <span className="text-sm font-medium text-gray-700">Calculator</span>
                <button
                    type="button"
                    onClick={onToggle}
                    onMouseDown={(e) => e.preventDefault()}
                    className="text-gray-400 hover:text-gray-600 transition-colors p-1"
                    aria-label="Close calculator"
                >
                    <svg
                        xmlns="http://www.w3.org/2000/svg"
                        className="h-5 w-5"
                        viewBox="0 0 20 20"
                        fill="currentColor"
                    >
                        <path
                            fillRule="evenodd"
                            d="M4.293 4.293a1 1 0 011.414 0L10 8.586l4.293-4.293a1 1 0 111.414 1.414L11.414 10l4.293 4.293a1 1 0 01-1.414 1.414L10 11.414l-4.293 4.293a1 1 0 01-1.414-1.414L8.586 10 4.293 5.707a1 1 0 010-1.414z"
                            clipRule="evenodd"
                        />
                    </svg>
                </button>
            </div>

            {/* Buttons grid */}
            <div className="grid grid-cols-4 gap-2">
                {buttons.flat().map((btn, index) => {
                    const isOperator = ['+', '-', '*', '/', '='].includes(btn);
                    const isSpecial = ['C', '←'].includes(btn);

                    return (
                        <button
                            key={index}
                            type="button"
                            onClick={() => onButtonClick(btn)}
                            onMouseDown={(e) => e.preventDefault()}
                            className={`
                px-3 py-2 rounded font-medium text-sm transition-all
                ${isOperator
                                    ? 'bg-blue-500 hover:bg-blue-600 text-white'
                                    : isSpecial
                                        ? 'bg-red-500 hover:bg-red-600 text-white'
                                        : 'bg-gray-200 hover:bg-gray-300 text-gray-800'
                                }
                active:scale-95 shadow-sm
              `}
                        >
                            {btn}
                        </button>
                    );
                })}
            </div>
        </div>
    );
}


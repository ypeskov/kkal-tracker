import { Link } from '@tanstack/react-router';
import { useEffect, useRef, useState } from 'react';
import { useTranslation } from 'react-i18next';
// HamburgerMenu.css imports removed - using Tailwind CSS

interface HamburgerMenuProps {
  onLogout?: () => void;
}

export default function HamburgerMenu({ onLogout }: HamburgerMenuProps) {
  const { t } = useTranslation();
  const [isOpen, setIsOpen] = useState(false);
  const menuRef = useRef<HTMLDivElement>(null);

  const toggleMenu = () => {
    setIsOpen(!isOpen);
  };

  const closeMenu = () => {
    setIsOpen(false);
  };

  // Close menu when clicking outside
  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (menuRef.current && !menuRef.current.contains(event.target as Node)) {
        closeMenu();
      }
    };

    if (isOpen) {
      document.addEventListener('mousedown', handleClickOutside);
    }

    return () => {
      document.removeEventListener('mousedown', handleClickOutside);
    };
  }, [isOpen]);

  // Close menu on route change
  const handleLinkClick = () => {
    closeMenu();
  };


  // Add keyboard support (ESC key)
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === 'Escape' && isOpen) {
        closeMenu();
      }
    };

    if (isOpen) {
      document.addEventListener('keydown', handleKeyDown);
    }

    return () => {
      document.removeEventListener('keydown', handleKeyDown);
    };
  }, [isOpen]);

  return (
    <div className="relative z-50" ref={menuRef}>
      <button
        className="flex flex-col justify-center items-center w-11 h-11 bg-transparent border-0 cursor-pointer p-2 rounded-lg transition-colors duration-200 hover:bg-black/5 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2"
        onClick={toggleMenu}
        aria-label="Toggle navigation menu"
        aria-expanded={isOpen}
      >
        <span className={`block w-6 h-0.5 bg-gray-800 mx-0 my-0.5 transition-all duration-300 ease-[cubic-bezier(0.23,1,0.320,1)] rounded-sm ${isOpen ? 'rotate-45 translate-x-1 translate-y-1' : ''
          }`}></span>
        <span className={`block w-6 h-0.5 bg-gray-800 mx-0 my-0.5 transition-all duration-300 ease-[cubic-bezier(0.23,1,0.320,1)] rounded-sm ${isOpen ? 'opacity-0 -translate-x-5' : ''
          }`}></span>
        <span className={`block w-6 h-0.5 bg-gray-800 mx-0 my-0.5 transition-all duration-300 ease-[cubic-bezier(0.23,1,0.320,1)] rounded-sm ${isOpen ? '-rotate-45 translate-x-1.5 -translate-y-1.5' : ''
          }`}></span>
      </button>

      <nav className={`fixed top-0 right-0 h-screen w-70 bg-white shadow-[-2px_0_20px_rgba(0,0,0,0.1)] transition-transform duration-300 ease-[cubic-bezier(0.23,1,0.320,1)] z-40 pt-5 ${isOpen ? 'translate-x-0' : 'translate-x-full'
        } md:w-80 lg:w-96`}>
        <div className="flex justify-end px-5 py-3 border-b border-gray-100 mb-5">
          <button
            className="bg-transparent border-0 text-2xl text-gray-400 cursor-pointer p-2 rounded-full w-10 h-10 flex items-center justify-center transition-all duration-200 hover:bg-gray-50 hover:text-gray-600 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2"
            onClick={closeMenu}
            aria-label="Close menu"
          >
            ✕
          </button>
        </div>
        <ul className="list-none p-0 m-0">
          <li className="border-b border-gray-100">
            <Link to="/" onClick={handleLinkClick} className="block py-4 px-6 no-underline text-gray-800 text-base font-medium transition-all duration-200 border-l-4 border-transparent hover:bg-gray-50 hover:border-l-blue-500 hover:text-blue-500">
              {t('nav.dashboard')}
            </Link>
          </li>
          <li className="border-b border-gray-100">
            <Link to="/food" onClick={handleLinkClick} className="block py-4 px-6 no-underline text-gray-800 text-base font-medium transition-all duration-200 border-l-4 border-transparent hover:bg-gray-50 hover:border-l-blue-500 hover:text-blue-500">
              {t('nav.foodList')}
            </Link>
          </li>
          <li className="border-b border-gray-100">
            <Link to="/reports" onClick={handleLinkClick} className="block py-4 px-6 no-underline text-gray-800 text-base font-medium transition-all duration-200 border-l-4 border-transparent hover:bg-gray-50 hover:border-l-blue-500 hover:text-blue-500">
              {t('nav.reports')}
            </Link>
          </li>
          <li className="border-b border-gray-100">
            <Link to="/ai-insights" onClick={handleLinkClick} className="block py-4 px-6 no-underline text-gray-800 text-base font-medium transition-all duration-200 border-l-4 border-transparent hover:bg-gray-50 hover:border-l-purple-500 hover:text-purple-500">
              {t('nav.aiInsights')}
            </Link>
          </li>
          <li className="border-b border-gray-100">
            <Link to="/profile" onClick={handleLinkClick} className="block py-4 px-6 no-underline text-gray-800 text-base font-medium transition-all duration-200 border-l-4 border-transparent hover:bg-gray-50 hover:border-l-blue-500 hover:text-blue-500">
              {t('nav.profile')}
            </Link>
          </li>
          <li className="border-b border-gray-100">
            <Link to="/settings" onClick={handleLinkClick} className="block py-4 px-6 no-underline text-gray-800 text-base font-medium transition-all duration-200 border-l-4 border-transparent hover:bg-gray-50 hover:border-l-blue-500 hover:text-blue-500">
              {t('nav.settings')}
            </Link>
          </li>
          {onLogout && (
            <li className="border-b border-gray-100">
              <button
                onClick={() => {
                  onLogout();
                  closeMenu();
                }}
                className="block w-full py-4 px-6 bg-transparent border-0 text-left text-red-600 text-base font-medium cursor-pointer transition-colors duration-200 border-l-4 border-transparent hover:bg-gray-50 hover:border-l-red-600"
              >
                {t('auth.logout')}
              </button>
            </li>
          )}
        </ul>
      </nav>

      {isOpen && <div className="fixed inset-0 bg-black/30 z-30 backdrop-blur-sm" onClick={closeMenu}></div>}
    </div>
  );
}
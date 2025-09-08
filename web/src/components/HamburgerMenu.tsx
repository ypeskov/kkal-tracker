import { useState, useRef, useEffect } from 'react';
import { Link } from '@tanstack/react-router';
import { useTranslation } from 'react-i18next';
import './HamburgerMenu.css';

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
    <div className="hamburger-menu" ref={menuRef}>
      <button 
        className={`hamburger-toggle ${isOpen ? 'open' : ''}`}
        onClick={toggleMenu}
        aria-label="Toggle navigation menu"
        aria-expanded={isOpen}
      >
        <span></span>
        <span></span>
        <span></span>
      </button>

      <nav className={`hamburger-nav ${isOpen ? 'open' : ''}`}>
        <div className="hamburger-nav-header">
          <button 
            className="hamburger-close-btn" 
            onClick={closeMenu}
            aria-label="Close menu"
          >
            ✕
          </button>
        </div>
        <ul className="hamburger-nav-list">
          <li>
            <Link to="/" onClick={handleLinkClick} className="hamburger-nav-link">
              {t('nav.dashboard')}
            </Link>
          </li>
          <li>
            <Link to="/food" onClick={handleLinkClick} className="hamburger-nav-link">
              {t('nav.foodList')}
            </Link>
          </li>
          <li>
            <Link to="/reports" onClick={handleLinkClick} className="hamburger-nav-link">
              {t('nav.reports')}
            </Link>
          </li>
          <li>
            <Link to="/settings" onClick={handleLinkClick} className="hamburger-nav-link">
              {t('nav.settings')}
            </Link>
          </li>
          {onLogout && (
            <li>
              <button 
                onClick={() => {
                  onLogout();
                  closeMenu();
                }} 
                className="hamburger-logout-btn"
              >
                {t('auth.logout')}
              </button>
            </li>
          )}
        </ul>
      </nav>

      {isOpen && <div className="hamburger-overlay" onClick={closeMenu}></div>}
    </div>
  );
}
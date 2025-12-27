'use client';

import { useState, useRef, useEffect } from 'react';
import { ChevronDownIcon, CheckIcon } from '@heroicons/react/24/outline';

export interface DropdownOption {
  value: string;
  label: string;
}

interface CustomDropdownProps {
  value: string;
  onChange: (value: string) => void;
  options: DropdownOption[];
  placeholder?: string;
  className?: string;
  disabled?: boolean;
}

export default function CustomDropdown({
  value,
  onChange,
  options,
  placeholder = 'Select an option',
  className = '',
  disabled = false,
}: CustomDropdownProps) {
  const [isOpen, setIsOpen] = useState(false);
  const dropdownRef = useRef<HTMLDivElement>(null);

  const selectedOption = options.find((opt) => opt.value === value);

  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (dropdownRef.current && !dropdownRef.current.contains(event.target as Node)) {
        setIsOpen(false);
      }
    };

    if (isOpen) {
      document.addEventListener('mousedown', handleClickOutside);
    }

    return () => {
      document.removeEventListener('mousedown', handleClickOutside);
    };
  }, [isOpen]);

  const handleSelect = (optionValue: string) => {
    onChange(optionValue);
    setIsOpen(false);
  };

  return (
    <div className={`relative ${className}`} ref={dropdownRef}>
      <button
        type="button"
        onClick={() => !disabled && setIsOpen(!isOpen)}
        disabled={disabled}
        className={`
          w-full px-4 py-2 bg-gray-700 border border-gray-600 rounded-md text-white
          focus:outline-none focus:ring-2 focus:ring-blue-500
          flex items-center justify-between
          ${disabled ? 'opacity-50 cursor-not-allowed' : 'hover:bg-gray-600 cursor-pointer'}
          transition-colors duration-200
        `}
      >
        <span className={selectedOption ? 'text-white' : 'text-gray-400'}>
          {selectedOption ? selectedOption.label : placeholder}
        </span>
        <ChevronDownIcon
          className={`h-5 w-5 text-gray-400 transition-transform duration-200 ${
            isOpen ? 'transform rotate-180' : ''
          }`}
        />
      </button>

      {isOpen && !disabled && (
        <div className="absolute z-50 w-full mt-1 bg-gray-800 border border-gray-700 rounded-md shadow-lg max-h-60 overflow-auto">
          <div className="py-1">
            {options.map((option) => (
              <button
                key={option.value}
                type="button"
                onClick={() => handleSelect(option.value)}
                className={`
                  w-full px-4 py-2 text-left text-sm
                  flex items-center justify-between
                  transition-colors duration-150
                  ${
                    value === option.value
                      ? 'bg-blue-600 text-white'
                      : 'text-gray-300 hover:bg-gray-700 hover:text-white'
                  }
                `}
              >
                <span>{option.label}</span>
                {value === option.value && (
                  <CheckIcon className="h-4 w-4 text-white" />
                )}
              </button>
            ))}
          </div>
        </div>
      )}
    </div>
  );
}



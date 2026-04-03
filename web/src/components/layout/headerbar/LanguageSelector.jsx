/*
Copyright (C) 2025 QuantumNous

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as
published by the Free Software Foundation, either version 3 of the
License, or (at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program. If not, see <https://www.gnu.org/licenses/>.

For commercial licensing, please contact support@quantumnous.com
*/

import React from 'react';
import { Button, Dropdown } from '@douyinfe/semi-ui';
import { Languages } from 'lucide-react';
import { CN, GB, FR, RU, JP, VN } from 'country-flag-icons/react/3x2';

const LanguageSelector = ({ currentLang, onLanguageChange, t }) => {
  return (
    <Dropdown
      position='bottomRight'
      render={
        <Dropdown.Menu className='!bg-semi-color-bg-overlay !border-semi-color-border !shadow-lg !rounded-lg'>
          {/* Language sorting: Order by English name (Chinese, English, French, Japanese, Russian) */}
          <Dropdown.Item
            onClick={() => onLanguageChange('zh')}
            className={`!flex !items-center !gap-2 !px-3 !py-1.5 !text-sm !text-semi-color-text-0 ${currentLang === 'zh' ? '!bg-semi-color-primary-light-default !font-semibold' : 'hover:!bg-semi-color-fill-1'}`}
          >
            <CN title='中文' className='!w-5 !h-auto' />
            <span>中文</span>
          </Dropdown.Item>
          <Dropdown.Item
            onClick={() => onLanguageChange('en')}
            className={`!flex !items-center !gap-2 !px-3 !py-1.5 !text-sm !text-semi-color-text-0 ${currentLang === 'en' ? '!bg-semi-color-primary-light-default !font-semibold' : 'hover:!bg-semi-color-fill-1'}`}
          >
            <GB title='English' className='!w-5 !h-auto' />
            <span>English</span>
          </Dropdown.Item>
          <Dropdown.Item
            onClick={() => onLanguageChange('fr')}
            className={`!flex !items-center !gap-2 !px-3 !py-1.5 !text-sm !text-semi-color-text-0 ${currentLang === 'fr' ? '!bg-semi-color-primary-light-default !font-semibold' : 'hover:!bg-semi-color-fill-1'}`}
          >
            <FR title='Français' className='!w-5 !h-auto' />
            <span>Français</span>
          </Dropdown.Item>
          <Dropdown.Item
            onClick={() => onLanguageChange('ja')}
            className={`!flex !items-center !gap-2 !px-3 !py-1.5 !text-sm !text-semi-color-text-0 ${currentLang === 'ja' ? '!bg-semi-color-primary-light-default !font-semibold' : 'hover:!bg-semi-color-fill-1'}`}
          >
            <JP title='日本語' className='!w-5 !h-auto' />
            <span>日本語</span>
          </Dropdown.Item>
          <Dropdown.Item
            onClick={() => onLanguageChange('ru')}
            className={`!flex !items-center !gap-2 !px-3 !py-1.5 !text-sm !text-semi-color-text-0 ${currentLang === 'ru' ? '!bg-semi-color-primary-light-default !font-semibold' : 'hover:!bg-semi-color-fill-1'}`}
          >
            <RU title='Русский' className='!w-5 !h-auto' />
            <span>Русский</span>
          </Dropdown.Item>
          <Dropdown.Item
            onClick={() => onLanguageChange('vi')}
            className={`!flex !items-center !gap-2 !px-3 !py-1.5 !text-sm !text-semi-color-text-0 ${currentLang === 'vi' ? '!bg-semi-color-primary-light-default !font-semibold' : 'hover:!bg-semi-color-fill-1'}`}
          >
            <VN title='Tiếng Việt' className='!w-5 !h-auto' />
            <span>Tiếng Việt</span>
          </Dropdown.Item>
        </Dropdown.Menu>
      }
    >
      <Button
        icon={<Languages size={18} />}
        aria-label={t('common.changeLanguage')}
        theme='borderless'
        type='tertiary'
        className='!p-1.5 !text-current focus:!bg-semi-color-fill-1 !rounded-lg !bg-semi-color-fill-0 hover:!bg-semi-color-fill-1 !w-8 !h-8'
      />
    </Dropdown>
  );
};

export default LanguageSelector;

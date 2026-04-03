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
import { Button } from '@douyinfe/semi-ui';
import { RefreshCw, Search } from 'lucide-react';

const DashboardHeader = ({
  getGreeting,
  greetingVisible,
  showSearchModal,
  refresh,
  loading,
  t,
}) => {
  const today = new Date();
  const dateStr = `${today.getFullYear()}-${String(today.getMonth() + 1).padStart(2, '0')}-${String(today.getDate()).padStart(2, '0')}`;

  return (
    <div className='flex items-center justify-between mb-5'>
      <div>
        <h2
          className='text-2xl font-semibold text-semi-color-text-0 transition-opacity duration-1000 ease-in-out'
          style={{ opacity: greetingVisible ? 1 : 0 }}
        >
          {getGreeting}
        </h2>
        <p className='text-sm text-semi-color-text-2 mt-1'>{dateStr}</p>
      </div>
      <div className='flex gap-2'>
        <Button
          type='tertiary'
          theme='light'
          icon={<Search size={16} />}
          onClick={showSearchModal}
          className='!rounded-lg'
        />
        <Button
          type='tertiary'
          theme='light'
          icon={<RefreshCw size={16} />}
          onClick={refresh}
          loading={loading}
          className='!rounded-lg'
        />
      </div>
    </div>
  );
};

export default DashboardHeader;

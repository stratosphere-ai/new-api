import React from 'react';
import { Button } from '@douyinfe/semi-ui';
import { Key, BarChart3, CreditCard, Store } from 'lucide-react';
import { useNavigate } from 'react-router-dom';
import { useTranslation } from 'react-i18next';

const QuickActions = () => {
  const navigate = useNavigate();
  const { t } = useTranslation();

  const actions = [
    {
      label: t('创建令牌'),
      icon: <Key size={14} />,
      to: '/console/token',
    },
    {
      label: t('查看日志'),
      icon: <BarChart3 size={14} />,
      to: '/console/log',
    },
    {
      label: t('充值余额'),
      icon: <CreditCard size={14} />,
      to: '/console/topup',
    },
    {
      label: t('模型广场'),
      icon: <Store size={14} />,
      to: '/pricing',
    },
  ];

  return (
    <div className='flex flex-wrap gap-2 mb-5'>
      {actions.map((action) => (
        <Button
          key={action.to}
          type='tertiary'
          theme='light'
          size='small'
          icon={action.icon}
          className='!rounded-lg'
          onClick={() => navigate(action.to)}
        >
          {action.label}
        </Button>
      ))}
    </div>
  );
};

export default QuickActions;

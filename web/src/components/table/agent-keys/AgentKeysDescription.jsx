import React from 'react';
import { Typography } from '@douyinfe/semi-ui';
import { Bot } from 'lucide-react';

const { Text } = Typography;

const AgentKeysDescription = ({ t }) => {
  return (
    <div className='flex flex-col md:flex-row justify-between items-start md:items-center gap-2 w-full'>
      <div className='flex items-center text-blue-500'>
        <Bot size={16} className='mr-2' />
        <Text>{t('Agent 密钥管理')}</Text>
      </div>
      <Text type='tertiary' size='small'>
        {t('创建 Agent 密钥供您的 AI 机器人自主购买额度')}
      </Text>
    </div>
  );
};

export default AgentKeysDescription;

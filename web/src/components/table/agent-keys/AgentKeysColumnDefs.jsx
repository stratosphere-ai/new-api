import React from 'react';
import { Tag, Button, Space, Tooltip, Popover, Typography } from '@douyinfe/semi-ui';
import { IconCopy, IconDelete, IconEdit } from '@douyinfe/semi-icons';

const { Text, Paragraph } = Typography;

export const getAgentKeysColumns = ({
  t,
  copyText,
  setEditingKey,
  setShowEdit,
  deleteAgentKey,
}) => {
  return [
    {
      title: t('名称'),
      dataIndex: 'name',
      key: 'name',
      width: 150,
    },
    {
      title: t('状态'),
      dataIndex: 'status',
      key: 'status',
      width: 80,
      render: (status) => {
        return status === 1 ? (
          <Tag color='green' shape='circle' size='small'>
            {t('已启用')}
          </Tag>
        ) : (
          <Tag color='red' shape='circle' size='small'>
            {t('已禁用')}
          </Tag>
        );
      },
    },
    {
      title: t('密钥'),
      dataIndex: 'key',
      key: 'key',
      width: 200,
      render: (key) => {
        const display = key ? `ak-${key}` : '';
        return (
          <Space>
            <Text copyable={false} style={{ fontFamily: 'monospace', fontSize: 13 }}>
              {display}
            </Text>
            {key && (
              <Button
                icon={<IconCopy />}
                size='small'
                type='tertiary'
                onClick={() => copyText(`ak-${key}`)}
              />
            )}
          </Space>
        );
      },
    },
    {
      title: t('允许模型'),
      dataIndex: 'models',
      key: 'models',
      width: 180,
      render: (models) => {
        if (!models) return <Text type='tertiary'>{t('全部模型')}</Text>;
        const list = models.split(',').filter(Boolean);
        if (list.length <= 2) {
          return list.map((m) => (
            <Tag key={m} size='small' style={{ marginRight: 2 }}>
              {m}
            </Tag>
          ));
        }
        return (
          <Popover
            content={
              <div style={{ maxWidth: 300 }}>
                {list.map((m) => (
                  <Tag key={m} size='small' style={{ margin: 2 }}>
                    {m}
                  </Tag>
                ))}
              </div>
            }
            position='top'
          >
            <span>
              <Tag size='small'>{list[0]}</Tag>
              <Tag size='small'>+{list.length - 1}</Tag>
            </span>
          </Popover>
        );
      },
    },
    {
      title: t('月预算'),
      dataIndex: 'monthly_budget',
      key: 'monthly_budget',
      width: 120,
      render: (budget) => {
        if (!budget || budget === 0) {
          return <Text type='tertiary'>{t('不限')}</Text>;
        }
        const usd = budget / 500000;
        return <Text>${usd.toFixed(2)}</Text>;
      },
    },
    {
      title: t('单次限额'),
      dataIndex: 'max_quota_per_request',
      key: 'max_quota_per_request',
      width: 120,
      render: (max) => {
        if (!max || max === 0) {
          return <Text type='tertiary'>{t('不限')}</Text>;
        }
        const usd = max / 500000;
        return <Text>${usd.toFixed(2)}</Text>;
      },
    },
    {
      title: t('默认有效期'),
      dataIndex: 'default_expire',
      key: 'default_expire',
      width: 100,
      render: (expire) => expire || '30d',
    },
    {
      title: t('最后使用'),
      dataIndex: 'last_used_at',
      key: 'last_used_at',
      width: 160,
      render: (val) => {
        if (!val) return <Text type='tertiary'>{t('从未')}</Text>;
        return new Date(val).toLocaleString();
      },
    },
    {
      title: t('操作'),
      key: 'operations',
      width: 120,
      fixed: 'right',
      render: (_, record) => {
        return (
          <Space>
            <Tooltip content={t('编辑')}>
              <Button
                icon={<IconEdit />}
                size='small'
                type='tertiary'
                onClick={() => {
                  setEditingKey(record);
                  setShowEdit(true);
                }}
              />
            </Tooltip>
            <Tooltip content={t('删除')}>
              <Button
                icon={<IconDelete />}
                size='small'
                type='danger'
                onClick={() => deleteAgentKey(record.id)}
              />
            </Tooltip>
          </Space>
        );
      },
    },
  ];
};

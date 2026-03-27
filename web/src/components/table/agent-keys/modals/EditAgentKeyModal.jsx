import React, { useEffect, useRef, useState } from 'react';
import {
  SideSheet,
  Form,
  Button,
  Card,
  Space,
  Typography,
  Banner,
  AutoComplete,
} from '@douyinfe/semi-ui';
import { useTranslation } from 'react-i18next';
import { API, showError, showSuccess, copy } from '../../../../helpers';
import { useIsMobile } from '../../../../hooks/common/useIsMobile';

const { Text } = Typography;

const getInitValues = () => ({
  name: '',
  models: '',
  monthly_budget: 0,
  max_quota_per_request: 0,
  default_expire: '30d',
  allow_ips: '',
});

const EditAgentKeyModal = ({ visible, editingKey, handleClose, refresh }) => {
  const { t } = useTranslation();
  const isMobile = useIsMobile();
  const formApiRef = useRef(null);
  const [submitting, setSubmitting] = useState(false);
  const [createdKey, setCreatedKey] = useState(null);
  const [provisionEndpoint, setProvisionEndpoint] = useState('');

  const isEdit = editingKey && editingKey.id !== undefined;

  useEffect(() => {
    if (visible && isEdit && formApiRef.current) {
      // Load existing key data
      formApiRef.current.setValues({
        name: editingKey.name || '',
        models: editingKey.models || '',
        monthly_budget: editingKey.monthly_budget || 0,
        max_quota_per_request: editingKey.max_quota_per_request || 0,
        default_expire: editingKey.default_expire || '30d',
        allow_ips:
          editingKey.allow_ips !== null && editingKey.allow_ips !== undefined
            ? editingKey.allow_ips
            : '',
      });
    }
    if (visible && !isEdit) {
      setCreatedKey(null);
    }
  }, [visible, editingKey]);

  const submit = async () => {
    const values = formApiRef.current.getValues();
    if (!values.name || values.name.trim() === '') {
      showError(t('请输入名称'));
      return;
    }

    setSubmitting(true);
    try {
      const body = {
        name: values.name.trim(),
        models: values.models
          ? values.models
              .split(',')
              .map((m) => m.trim())
              .filter(Boolean)
          : [],
        monthly_budget: parseInt(values.monthly_budget) || 0,
        max_quota_per_request: parseInt(values.max_quota_per_request) || 0,
        default_expire: values.default_expire || '30d',
        allow_ips: values.allow_ips || '',
      };

      let res;
      if (isEdit) {
        res = await API.put(`/api/agent/keys/${editingKey.id}`, body);
      } else {
        res = await API.post('/api/agent/keys/', body);
      }

      if (res.data.success) {
        if (!isEdit && res.data.data) {
          setCreatedKey(res.data.data.key);
          setProvisionEndpoint(res.data.data.provision_endpoint);
          showSuccess(t('创建成功！请复制密钥，它只会显示一次。'));
        } else {
          showSuccess(t('操作成功完成！'));
          handleClose();
        }
        refresh();
      } else {
        showError(res.data.message);
      }
    } catch (e) {
      showError(e.message);
    }
    setSubmitting(false);
  };

  const handleCopyKey = async () => {
    if (createdKey && (await copy(createdKey))) {
      showSuccess(t('已复制到剪贴板！'));
    }
  };

  const budgetSuggestions = [
    { value: 0, label: t('不限') },
    { value: 500000, label: '$1' },
    { value: 2500000, label: '$5' },
    { value: 5000000, label: '$10' },
    { value: 25000000, label: '$50' },
    { value: 50000000, label: '$100' },
  ];

  return (
    <SideSheet
      title={isEdit ? t('编辑 Agent 密钥') : t('创建 Agent 密钥')}
      visible={visible}
      onCancel={handleClose}
      placement={isEdit ? 'right' : 'left'}
      width={isMobile ? '100%' : 560}
      footer={
        createdKey ? null : (
          <div className='flex justify-end gap-2'>
            <Button onClick={handleClose}>{t('取消')}</Button>
            <Button type='primary' loading={submitting} onClick={submit}>
              {isEdit ? t('保存') : t('创建')}
            </Button>
          </div>
        )
      }
    >
      {createdKey ? (
        <div className='flex flex-col gap-4'>
          <Banner
            type='success'
            description={t('Agent 密钥创建成功！请立即复制，此密钥只会显示一次。')}
          />

          <Card title={t('Agent 密钥')}>
            <div className='flex items-center gap-2'>
              <Text
                copyable={false}
                style={{
                  fontFamily: 'monospace',
                  fontSize: 14,
                  wordBreak: 'break-all',
                  flex: 1,
                }}
              >
                {createdKey}
              </Text>
              <Button type='primary' onClick={handleCopyKey}>
                {t('复制密钥')}
              </Button>
            </div>
          </Card>

          <Card title={t('接入地址')}>
            <Text
              copyable
              style={{ fontFamily: 'monospace', fontSize: 13 }}
            >
              {provisionEndpoint}
            </Text>
          </Card>

          <Card title={t('使用示例')}>
            <div
              style={{
                background: 'var(--semi-color-fill-0)',
                padding: 12,
                borderRadius: 6,
                fontFamily: 'monospace',
                fontSize: 12,
                whiteSpace: 'pre-wrap',
                wordBreak: 'break-all',
              }}
            >
              {`curl -X POST ${provisionEndpoint} \\
  -H "Authorization: Bearer ${createdKey}" \\
  -H "Content-Type: application/json" \\
  -d '{}'`}
            </div>
          </Card>

          <Button block onClick={handleClose}>
            {t('完成')}
          </Button>
        </div>
      ) : (
        <Form
          initValues={getInitValues()}
          getFormApi={(api) => (formApiRef.current = api)}
          allowEmpty
          layout='vertical'
        >
          <Card
            title={t('基本信息')}
            style={{ marginBottom: 16 }}
          >
            <Form.Input
              field='name'
              label={t('名称')}
              placeholder={t('例如：龙虾机器人')}
              rules={[{ required: true, message: t('请输入名称') }]}
            />

            <Form.TextArea
              field='models'
              label={t('允许模型')}
              placeholder={t('模型名称用逗号分隔，留空则允许所有模型。例如: gpt-4o,claude-sonnet-4-20250514')}
              autosize
              rows={2}
            />
          </Card>

          <Card
            title={t('预算控制')}
            style={{ marginBottom: 16 }}
          >
            <Form.Select
              field='monthly_budget'
              label={t('月预算上限')}
              optionList={budgetSuggestions.map((s) => ({
                value: s.value,
                label: `${s.label}${s.value > 0 ? ` (${s.value} quota)` : ''}`,
              }))}
              style={{ width: '100%' }}
            />
            <Text type='tertiary' size='small'>
              {t('预算内自动批准，超出需要邮件确认。0 表示不限制。')}
            </Text>

            <Form.Select
              field='max_quota_per_request'
              label={t('单次限额')}
              style={{ width: '100%', marginTop: 12 }}
              optionList={budgetSuggestions.map((s) => ({
                value: s.value,
                label: `${s.label}${s.value > 0 ? ` (${s.value} quota)` : ''}`,
              }))}
            />
            <Text type='tertiary' size='small'>
              {t('每次购买的最大额度。0 表示不限制。')}
            </Text>
          </Card>

          <Card title={t('高级设置')}>
            <Form.Select
              field='default_expire'
              label={t('默认 Token 有效期')}
              optionList={[
                { value: '1h', label: '1 ' + t('小时') },
                { value: '1d', label: '1 ' + t('天') },
                { value: '7d', label: '7 ' + t('天') },
                { value: '30d', label: '30 ' + t('天') },
                { value: '90d', label: '90 ' + t('天') },
                { value: '365d', label: '365 ' + t('天') },
              ]}
              style={{ width: '100%' }}
            />

            <Form.TextArea
              field='allow_ips'
              label={t('IP 白名单')}
              placeholder={t('每行一个 IP 地址，留空则不限制')}
              autosize
              rows={2}
              style={{ marginTop: 12 }}
            />
          </Card>
        </Form>
      )}
    </SideSheet>
  );
};

export default EditAgentKeyModal;

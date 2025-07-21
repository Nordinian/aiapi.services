import React, { useEffect, useState, useRef } from 'react';
import {
  Button,
  Form,
  Row,
  Col,
  Typography,
  Spin,
  Divider,
  Tabs,
  TabPane,
} from '@douyinfe/semi-ui';
const { Text } = Typography;
import {
  API,
  removeTrailingSlash,
  showError,
  showSuccess,
  verifyJSON,
} from '../../../helpers';
import { useTranslation } from 'react-i18next';

export default function SettingsPaymentGateway(props) {
  const { t } = useTranslation();
  const [loading, setLoading] = useState(false);
  const [inputs, setInputs] = useState({
    PayAddress: '',
    EpayId: '',
    EpayKey: '',
    Price: 7.3,
    MinTopUp: 1,
    TopupGroupRatio: '',
    CustomCallbackAddress: '',
    PayMethods: '',
    StripePriceID: '',
    StripeSecretKey: '',
    StripeWebhookSecret: '',
  });
  const [originInputs, setOriginInputs] = useState({});
  const formApiRef = useRef(null);

  useEffect(() => {
    if (props.options && formApiRef.current) {
      const currentInputs = {
        PayAddress: props.options.PayAddress || '',
        EpayId: props.options.EpayId || '',
        EpayKey: props.options.EpayKey || '',
        Price: props.options.Price !== undefined ? parseFloat(props.options.Price) : 7.3,
        MinTopUp: props.options.MinTopUp !== undefined ? parseFloat(props.options.MinTopUp) : 1,
        TopupGroupRatio: props.options.TopupGroupRatio || '',
        CustomCallbackAddress: props.options.CustomCallbackAddress || '',
        PayMethods: props.options.PayMethods || '',
        StripePriceID: props.options.StripePriceID || '',
        StripeSecretKey: props.options.StripeSecretKey || '',
        StripeWebhookSecret: props.options.StripeWebhookSecret || '',
      };
      setInputs(currentInputs);
      setOriginInputs({ ...currentInputs });
      formApiRef.current.setValues(currentInputs);
    }
  }, [props.options]);

  const handleFormChange = (values) => {
    setInputs(values);
  };

  const submitSettings = async (keys) => {
    setLoading(true);
    try {
      const options = [];
      for (const key of keys) {
        if (originInputs[key] !== inputs[key]) {
          // Check for JSON validity if it's one of these fields
          if (
            (key === 'TopupGroupRatio' || key === 'PayMethods') &&
            !verifyJSON(inputs[key])
          ) {
            showError(t(`${key} 不是合法的 JSON 字符串`));
            setLoading(false);
            return;
          }
          options.push({ key, value: inputs[key].toString() });
        }
      }

      if (options.length === 0) {
        showSuccess(t('设置未更改'));
        setLoading(false);
        return;
      }

      const requestQueue = options.map((opt) =>
        API.put('/api/option/', {
          key: opt.key,
          value: removeTrailingSlash(opt.value),
        }),
      );

      const results = await Promise.all(requestQueue);

      const errorResults = results.filter((res) => !res.data.success);
      if (errorResults.length > 0) {
        errorResults.forEach((res) => {
          showError(res.data.message);
        });
      } else {
        showSuccess(t('更新成功'));
        // Update local storage of original values
        setOriginInputs({ ...inputs });
        props.refresh && props.refresh();
      }
    } catch (error) {
      showError(t('更新失败'));
    }
    setLoading(false);
  };

  const epayKeys = [
    'PayAddress',
    'EpayId',
    'EpayKey',
    'CustomCallbackAddress',
    'Price',
    'MinTopUp',
  ];
  const stripeKeys = ['StripePriceID', 'StripeSecretKey', 'StripeWebhookSecret'];
  const generalKeys = ['TopupGroupRatio', 'PayMethods'];

  return (
    <Spin spinning={loading}>
      <Form
        initValues={inputs}
        onValueChange={handleFormChange}
        getFormApi={(api) => (formApiRef.current = api)}
      >
        <Tabs type='line' defaultActiveKey='1'>
          <TabPane tab={t('易支付')} itemKey='1'>
            <Row>
              <Col span={24}>
                <Form.Input
                  field='PayAddress'
                  label={t('支付地址')}
                  placeholder={t('例如：https://yourdomain.com')}
                />
              </Col>
              <Col span={24}>
                <Form.Input
                  field='EpayId'
                  label={t('易支付商户ID')}
                  placeholder={t('例如：0001')}
                />
              </Col>
              <Col span={24}>
                <Form.Input
                  field='EpayKey'
                  label={t('易支付商户密钥')}
                  placeholder={t('敏感信息不会发送到前端显示')}
                  type='password'
                />
              </Col>
              <Col span={24}>
                <Form.Input
                  field='CustomCallbackAddress'
                  label={t('回调地址')}
                  placeholder={t('例如：https://yourdomain.com')}
                />
              </Col>
              <Col span={24}>
                <Form.InputNumber
                  field='Price'
                  precision={2}
                  label={t('充值价格（x元/美金）')}
                  placeholder={t('例如：7，就是7元/美金')}
                />
              </Col>
              <Col span={24}>
                <Form.InputNumber
                  field='MinTopUp'
                  label={t('最低充值美元数量')}
                  placeholder={t('例如：2，就是最低充值2$')}
                />
              </Col>
            </Row>
            <Button
              type='primary'
              style={{ marginTop: 16 }}
              onClick={() => submitSettings(epayKeys)}
            >
              {t('保存易支付设置')}
            </Button>
          </TabPane>
          <TabPane tab='Stripe' itemKey='2'>
            <Row>
              <Col span={24}>
                <Form.Input
                  field='StripePriceID'
                  label={t('Stripe 价格 ID')}
                  placeholder={t('Stripe 的价格 ID，例如 price_xxx')}
                />
              </Col>
              <Col span={24}>
                <Form.Input
                  field='StripeSecretKey'
                  label={t('Stripe 密钥')}
                  placeholder={t('sk_test_... 或 sk_live_...')}
                  type='password'
                />
              </Col>
              <Col span={24}>
                <Form.Input
                  field='StripeWebhookSecret'
                  label={t('Stripe Webhook 密钥')}
                  placeholder={t('whsec_...')}
                  type='password'
                />
              </Col>
            </Row>
            <Button
              type='primary'
              style={{ marginTop: 16 }}
              onClick={() => submitSettings(stripeKeys)}
            >
              {t('保存 Stripe 设置')}
            </Button>
          </TabPane>
          <TabPane tab={t('通用设置')} itemKey='3'>
            <Row>
              <Col span={24}>
                <Form.TextArea
                  field='TopupGroupRatio'
                  label={t('充值分组倍率')}
                  placeholder={t('为一个 JSON 文本，键为组名称，值为倍率')}
                  autosize
                />
              </Col>
              <Col span={24}>
                <Form.TextArea
                  field='PayMethods'
                  label={t('充值方式设置')}
                  placeholder={t('为一个 JSON 文本')}
                  autosize
                />
              </Col>
            </Row>
            <Button
              type='primary'
              style={{ marginTop: 16 }}
              onClick={() => submitSettings(generalKeys)}
            >
              {t('保存通用设置')}
            </Button>
          </TabPane>
        </Tabs>
      </Form>
    </Spin>
  );
} 
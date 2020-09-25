import { Box, Button, Grid } from "@material-ui/core";
import { AutoCompleteMultiValuesFreeSolo } from "forms/Final/autoComplete";
import { ValidatorArrayOfIsValidHostInCertificate } from "forms/validator";
import React from "react";
import { Field, FieldRenderProps, Form } from "react-final-form";
import { connect } from "react-redux";
import { RootState } from "reducers";
import { TDispatchProp } from "types";
import { CertificateFormType } from "types/certificate";
import { KPanel } from "widgets/KPanel";
import { Body, Caption } from "widgets/Label";
import { Prompt } from "widgets/Prompt";
import sc from "../../utils/stringConstants";
import { FormValueToReudxStoreListener } from "tutorials/formValueToReudxStoreListener";
import { CERTIFICATE_FORM_ID } from "forms/formIDs";
import { FormDataPreview } from "forms/Final/util";
import { stringArrayTrimAndToLowerCaseParse } from "forms/normalizer";

const mapStateToProps = (state: RootState) => {
  return {};
};

interface OwnProps {
  isEdit?: boolean;
  onSubmit: any;
  initialValues: CertificateFormType;
}

export interface Props extends ReturnType<typeof mapStateToProps>, TDispatchProp, OwnProps {}

class CertificateFormRaw extends React.PureComponent<Props> {
  public render() {
    const { onSubmit, initialValues, isEdit } = this.props;
    return (
      <Form onSubmit={onSubmit} initialValues={initialValues} keepDirtyOnReinitialize>
        {(props) => {
          const { values, dirty, submitting, handleSubmit } = props;

          return (
            <form onSubmit={handleSubmit} tutorial-anchor-id="certificate-form" id="certificate-form">
              <Box p={2}>
                <FormValueToReudxStoreListener values={values} form={CERTIFICATE_FORM_ID} />
                <Prompt when={dirty && !submitting} message={sc.CONFIRM_LEAVE_WITHOUT_SAVING} />
                <KPanel
                  content={
                    <Box p={2}>
                      <Grid container spacing={2}>
                        <Grid item md={12}>
                          <Body>{sc.CERT_DNS01}</Body>
                          <Caption>{sc.CERT_DNS01_DESC}</Caption>
                        </Grid>
                        <Grid item md={12}>
                          <Field
                            render={(props: FieldRenderProps<string[]>) => (
                              <AutoCompleteMultiValuesFreeSolo<string> {...props} options={[]} />
                            )}
                            label="Domains"
                            name="domains"
                            validate={ValidatorArrayOfIsValidHostInCertificate}
                            parse={stringArrayTrimAndToLowerCaseParse}
                            id="certificate-domains"
                            placeholder="e.g. foo.com; *.foo.bar.com"
                          />
                        </Grid>
                      </Grid>
                    </Box>
                  }
                />

                <FormDataPreview />

                <Box pt={2}>
                  <Button id="save-certificate-button" type="submit" color="primary" variant="contained">
                    {isEdit ? "Update" : "Create"}
                  </Button>
                </Box>
              </Box>
            </form>
          );
        }}
      </Form>
    );
  }
}

export const CertificateForm = connect(mapStateToProps)(CertificateFormRaw);

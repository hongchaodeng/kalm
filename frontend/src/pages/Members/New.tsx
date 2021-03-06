import { Box, createStyles, Grid, Theme, withStyles, WithStyles } from "@material-ui/core";
import { setSuccessNotificationAction } from "actions/notification";
import { createRoleBindingsAction } from "actions/user";
import { push } from "connected-react-router";
import { MemberForm } from "forms/Member";
import { withNamespace, WithNamespaceProps } from "hoc/withNamespace";
import { ApplicationSidebar } from "pages/Application/ApplicationSidebar";
import { BasePage } from "pages/BasePage";
import React from "react";
import { connect } from "react-redux";
import { withRouter } from "react-router-dom";
import { RootState } from "reducers";
import { TDispatchProp } from "types";
import { newEmptyRoleBinding, RoleBinding } from "types/member";
import { Namespaces } from "widgets/Namespaces";

const styles = (theme: Theme) => createStyles({});

const mapStateToProps = (state: RootState) => {
  return {};
};

interface Props
  extends WithStyles<typeof styles>,
    ReturnType<typeof mapStateToProps>,
    TDispatchProp,
    WithNamespaceProps {}

class MemberNewPageRaw extends React.PureComponent<Props> {
  private onSubmit = async (values: RoleBinding) => {
    const { dispatch, activeNamespaceName } = this.props;
    values.namespace = this.isClusterLevel() ? "kalm-system" : activeNamespaceName;
    await dispatch(createRoleBindingsAction(values));
    await dispatch(setSuccessNotificationAction("Successfully create role binding"));
    if (this.isClusterLevel()) {
      await dispatch(push("/cluster/members"));
    } else {
      await dispatch(push("/applications/" + activeNamespaceName + "/members"));
    }
  };

  private isClusterLevel() {
    const {
      location: { pathname },
    } = this.props;
    return pathname.startsWith("/cluster/members") || pathname.startsWith("/applications/kalm-system");
  }

  public render() {
    const isClusterLevel = this.isClusterLevel();
    return (
      <BasePage
        secondHeaderLeft={isClusterLevel ? null : <Namespaces />}
        leftDrawer={isClusterLevel ? null : <ApplicationSidebar />}
      >
        <Box p={2}>
          <Grid container spacing={2}>
            <Grid item xs={8} sm={8} md={8}>
              <MemberForm
                initial={newEmptyRoleBinding(isClusterLevel)}
                onSubmit={this.onSubmit}
                isClusterLevel={isClusterLevel}
              />
            </Grid>
          </Grid>
        </Box>
      </BasePage>
    );
  }
}

export const MemberNewPage = withStyles(styles)(withNamespace(connect(mapStateToProps)(withRouter(MemberNewPageRaw))));

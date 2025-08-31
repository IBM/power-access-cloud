<#import "template.ftl" as layout>
<#import "social-providers.ftl" as identityProviders>

<@layout.registrationLayout displayMessage=false; section>
    <!-- template: login.ftl -->

    <#if section == "header">
        ${msg("loginAccountTitle")}

    <#elseif section == "socialProviders">
        <#if social.providers?? && social.providers?has_content>
            <@identityProviders.show social=social/>
        </#if>
    </#if>
</@layout.registrationLayout>

package session

import (
	"net/http"

	"github.com/supertokens/supertokens-golang/errors"
	"github.com/supertokens/supertokens-golang/recipe/session/api"
	"github.com/supertokens/supertokens-golang/recipe/session/models"
	"github.com/supertokens/supertokens-golang/supertokens"
)

const RECIPE_ID = "session"

var r *models.SessionRecipe = nil

func MakeRecipe(recipeId string, appInfo supertokens.NormalisedAppinfo, config *models.TypeInput) (models.SessionRecipe, error) {
	querierInstance, querierError := supertokens.GetNewQuerierInstanceOrThrowError(recipeId)
	if querierError != nil {
		return models.SessionRecipe{}, querierError
	}
	recipeModuleInstance := supertokens.MakeRecipeModule(recipeId, appInfo, HandleAPIRequest, GetAllCORSHeaders, GetAPIsHandled)

	// TODO: you need to pass a pointer to r right? instead of r itself. Cause when this is called
	// r is nil.
	verifiedConfig, configError := validateAndNormaliseUserInput(r, appInfo, config)
	if configError != nil {
		return models.SessionRecipe{}, configError
	}
	recipeImplementation := MakeRecipeImplementation(*querierInstance, verifiedConfig)

	return models.SessionRecipe{
		RecipeModule: recipeModuleInstance,
		Config:       verifiedConfig,
		RecipeImpl:   verifiedConfig.Override.Functions(recipeImplementation),
		APIImpl:      verifiedConfig.Override.APIs(api.MakeAPIImplementation()),
	}, nil
}

func GetInstanceOrThrowError() (*models.SessionRecipe, error) {
	if r != nil {
		return r, nil
	}
	return nil, errors.BadInputError{Msg: "Initialisation not done. Did you forget to call the init function?"}
}

func RecipeInit(config models.TypeInput) supertokens.RecipeListFunction {
	return func(appInfo supertokens.NormalisedAppinfo) (*supertokens.RecipeModule, error) {
		if r == nil {
			recipe, err := MakeRecipe(RECIPE_ID, appInfo, &config)
			if err != nil {
				return nil, err
			}
			r = &recipe
			return &r.RecipeModule, nil
		}
		return nil, errors.BadInputError{Msg: "Session recipe has already been initialised. Please check your code for bugs."}
	}
}

// Implement RecipeModule

func GetAPIsHandled() ([]supertokens.APIHandled, error) {
	refreshAPIPathNormalised, err := supertokens.NewNormalisedURLPath(refreshAPIPath)
	if err != nil {
		return nil, err
	}
	signoutAPIPathNormalised, err := supertokens.NewNormalisedURLPath(signoutAPIPath)
	if err != nil {
		return nil, err
	}
	return []supertokens.APIHandled{{
		Method:                 "post",
		PathWithoutAPIBasePath: *refreshAPIPathNormalised,
		ID:                     refreshAPIPath,
		Disabled:               r.APIImpl.RefreshPOST == nil,
	}, {
		Method:                 "post",
		PathWithoutAPIBasePath: *signoutAPIPathNormalised,
		ID:                     signoutAPIPath,
		Disabled:               r.APIImpl.SignOutPOST == nil,
	}}, nil
}

func HandleAPIRequest(id string, req *http.Request, res http.ResponseWriter, theirhandler http.HandlerFunc, _ supertokens.NormalisedURLPath, _ string) error {
	options := models.APIOptions{
		Config:               r.Config,
		RecipeID:             r.RecipeModule.GetRecipeID(),
		RecipeImplementation: r.RecipeImpl,
		Req:                  req,
		Res:                  res,
		OtherHandler:         theirhandler,
	}
	if id == refreshAPIPath {
		api.HandleRefreshAPI(r.APIImpl, options)
	} else {
		return api.SignOutAPI(r.APIImpl, options)
	}
	return nil
}

func GetAllCORSHeaders() []string {
	return getCORSAllowedHeaders()
}
